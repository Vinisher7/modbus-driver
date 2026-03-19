package database

import (
	"context"
	"database/sql"
	"fmt"
	"modbus-driver/internal/config"
	"modbus-driver/internal/models"

	_ "github.com/microsoft/go-mssqldb"
)

func Connect(cfg *config.Config) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"sqlserver://%s:%s@%s:%d?database=%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBServer, cfg.DBPort, cfg.DBName,
	)
	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err = db.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}

const query = `
SELECT
    CAST(d.id AS VARCHAR(36))            AS device_id,
    d.name                               AS device_name,
    CAST(md.id AS VARCHAR(36))           AS modbus_device_id,
    md.host,
    md.port,
    md.unit_id,
    CAST(mf.id AS VARCHAR(36))           AS fetch_id,
    mf.initial_address,
    mf.quantity,
    mf.function_code,
    mtf.position,
    CAST(mt.id AS VARCHAR(36))           AS tag_id,
    mt.name                              AS tag_name,
    mt.data_type,
    mt.operation,
    mt.operation_value
FROM devices d
JOIN modbus_devices   md   ON md.id_device        = d.id
JOIN modbus_fetchs    mf   ON mf.id_modbus_device  = md.id
JOIN modbus_tag_fetch mtf  ON mtf.id_modbus_fetch  = mf.id
JOIN modbus_tags      mt   ON mt.id               = mtf.id_modbus_tag
WHERE d.protocol = 'modbus'
ORDER BY d.id, mf.id, mt.id, mtf.position
`

// LoadDevices executa a query e monta a estrutura hierárquica de devices
func LoadDevices(ctx context.Context, db *sql.DB) ([]models.Device, error) {
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("query devices: %w", err)
	}
	defer rows.Close()

	deviceMap := map[string]*models.Device{}
	deviceOrder := []string{}
	fetchMap := map[string]*models.Fetch{}
	tagMap := map[string]*models.Tag{} // novo

	for rows.Next() {
		var r models.DeviceRow
		var (
			rawDeviceID       any
			rawModbusDeviceID any
			rawFetchID        any
			rawTagID          any
		)

		if err := rows.Scan(
			&rawDeviceID, &r.DeviceName, &rawModbusDeviceID,
			&r.Host, &r.Port, &r.UnitID,
			&rawFetchID, &r.InitialAddress, &r.Quantity, &r.FunctionCode,
			&r.Position, &rawTagID, &r.TagName, &r.DataType,
			&r.Operation, &r.OperationValue,
		); err != nil {
			return nil, err
		}

		r.DeviceID = scanUUID(rawDeviceID)
		r.ModbusDeviceID = scanUUID(rawModbusDeviceID)
		r.FetchID = scanUUID(rawFetchID)
		r.TagID = scanUUID(rawTagID)

		// device
		if _, ok := deviceMap[r.DeviceID]; !ok {
			deviceMap[r.DeviceID] = &models.Device{
				DeviceID:       r.DeviceID,
				DeviceName:     r.DeviceName,
				ModbusDeviceID: r.ModbusDeviceID,
				Host:           r.Host,
				Port:           r.Port,
				UnitID:         r.UnitID,
			}
			deviceOrder = append(deviceOrder, r.DeviceID)
		}

		// fetch
		fetchKey := r.DeviceID + ":" + r.FetchID
		if _, ok := fetchMap[fetchKey]; !ok {
			f := &models.Fetch{
				ID:             r.FetchID,
				InitialAddress: r.InitialAddress,
				Quantity:       r.Quantity,
				FunctionCode:   r.FunctionCode,
			}
			fetchMap[fetchKey] = f
			deviceMap[r.DeviceID].Fetches = append(deviceMap[r.DeviceID].Fetches, *f)
		}

		// tag — agrupa positions da mesma tag
		tagKey := fetchKey + ":" + r.TagID
		if _, ok := tagMap[tagKey]; !ok {
			tag := &models.Tag{
				ID:             r.TagID,
				Name:           r.TagName,
				DataType:       r.DataType,
				Operation:      r.Operation,
				OperationValue: r.OperationValue,
				Positions:      []int{},
			}
			tagMap[tagKey] = tag

			// adiciona a tag no fetch correto do device
			dev := deviceMap[r.DeviceID]
			for i := range dev.Fetches {
				if dev.Fetches[i].ID == r.FetchID {
					dev.Fetches[i].Tags = append(dev.Fetches[i].Tags, *tag)
					break
				}
			}
		}

		// sempre adiciona a position (pode ser a 2ª word)
		tagMap[tagKey].Positions = append(tagMap[tagKey].Positions, r.Position)

		// sincroniza o slice de volta no fetch (pointer no map mas value no slice)
		dev := deviceMap[r.DeviceID]
		for i := range dev.Fetches {
			if dev.Fetches[i].ID == r.FetchID {
				for j := range dev.Fetches[i].Tags {
					if dev.Fetches[i].Tags[j].ID == r.TagID {
						dev.Fetches[i].Tags[j].Positions = tagMap[tagKey].Positions
					}
				}
			}
		}
	}

	result := make([]models.Device, 0, len(deviceOrder))
	for _, id := range deviceOrder {
		result = append(result, *deviceMap[id])
	}
	return result, nil
}

func scanUUID(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case []byte:
		// uniqueidentifier do SQL Server vem em little-endian nos 3 primeiros grupos
		if len(val) == 16 {
			return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
				reorderBytes(val[0:4]),
				reorderBytes(val[4:6]),
				reorderBytes(val[6:8]),
				val[8:10],
				val[10:16],
			)
		}
		return string(val)
	}
	return ""
}

func reorderBytes(b []byte) []byte {
	c := make([]byte, len(b))
	copy(c, b)
	for i, j := 0, len(c)-1; i < j; i, j = i+1, j-1 {
		c[i], c[j] = c[j], c[i]
	}
	return c
}
