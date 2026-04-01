package database

import (
	"context"
	"fmt"
	"modbus-driver/internal/config/logger"
	"modbus-driver/internal/models"
	"os"
	"time"

	"github.com/jackc/pgx/v5"
	"go.uber.org/zap"
)

var (
	POSTGRES_USER     = "POSTGRES_USER"
	POSTGRES_PASSWORD = "POSTGRES_PASSWORD"
	POSTGRES_SERVER   = "POSTGRES_SERVER"
	POSTGRES_PORT     = "POSTGRES_PORT"
	POSTGRES_NAME     = "POSTGRES_DB"
)

func NewPostgresClient(ctx context.Context) (db *pgx.Conn, err error) {
	logger.Info("Init NewPostgresClient", zap.String("journey", "database"))

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	connString := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		os.Getenv(POSTGRES_USER),
		os.Getenv(POSTGRES_PASSWORD),
		os.Getenv(POSTGRES_SERVER),
		os.Getenv(POSTGRES_PORT),
		os.Getenv(POSTGRES_NAME),
	)

	db, err = pgx.Connect(ctx, connString)
	if err != nil {
		logger.Error("Connect func returned an error", err, zap.String("journey", "database"))
		return nil, fmt.Errorf("error connecting with postgres: %w", err)
	}

	logger.Info("NewPostgresClient executed successfully", zap.String("journey", "database"))

	return db, nil
}

func TestPostgresConnection(ctx context.Context, db *pgx.Conn) (err error) {
	logger.Info("Init TestPostgresConnection", zap.String("journey", "database"))

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	err = db.Ping(ctx)
	if err != nil {
		logger.Error("Ping func returned an error", err, zap.String("journey", "database"))
		return fmt.Errorf("error testing a connection with postgres: %w", err)
	}

	logger.Info("TestPostgresConnection executed successfully", zap.String("journey", "database"))

	return nil
}

const query = `
		SELECT
			CAST(d.id AS VARCHAR(36)) AS device_id,
			d.device_name,
			CAST(md.id AS VARCHAR(36)) AS modbus_device_id,
			md.host,
			md.port,
			md.unit_id,
			CAST(mf.id AS VARCHAR(36)) AS fetch_id,
			mf.initial_address,
			mf.quantity,
			mf.function_code,
			mtf.position,
			CAST(mt.id AS VARCHAR(36)) AS tag_id,
			mt.tag_name,
			mt.data_type,
			mt.operation,
			mt.operation_value
		FROM devices d
		JOIN modbus_devices md ON md.id_device = d.id
		JOIN modbus_fetchs mf ON mf.id_modbus_device = md.id
		JOIN modbus_tag_fetchs mtf ON mtf.id_modbus_fetch = mf.id
		JOIN modbus_tags mt ON mt.id = mtf.id_modbus_tag
		WHERE d.protocol = 'modbus'
		ORDER BY d.id, mf.id, mt.id, mtf.position
	`

func LoadDevices(ctx context.Context, db *pgx.Conn) ([]models.Device, error) {
	logger.Info("Init LoadDevices", zap.String("journey", "database"))

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	rows, err := db.Query(ctx, query)
	if err != nil {
		logger.Error("Query func returned an error", err, zap.String("journey", "database"))
		return nil, fmt.Errorf("error querying database: %w", err)
	}
	defer rows.Close()

	devicesRows, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.DeviceRow])
	if err != nil {
		logger.Error("CollectRows func returned an error", err, zap.String("journey", "database"))
		return nil, fmt.Errorf("error collecting rows: %w", err)
	}

	if len(devicesRows) == 0 {
		logger.Info("No devices found", zap.String("journey", "database"))
		return []models.Device{}, nil
	}

	deviceMap := make(map[string]*models.Device)
	fetchMap := make(map[string]*models.Fetch)
	tagMap := make(map[string]*models.Tag)

	var deviceOrder []string

	for _, r := range devicesRows {

		dev, exists := deviceMap[r.DeviceID]
		if !exists {
			dev = &models.Device{
				DeviceID:       r.DeviceID,
				DeviceName:     r.DeviceName,
				ModbusDeviceID: r.ModbusDeviceID,
				Host:           r.Host,
				Port:           r.Port,
				UnitID:         r.UnitID,
				Fetches:        make([]*models.Fetch, 0),
			}
			deviceMap[r.DeviceID] = dev
			deviceOrder = append(deviceOrder, r.DeviceID)
		}

		fetchKey := r.DeviceID + ":" + r.FetchID
		fetch, fExists := fetchMap[fetchKey]
		if !fExists {
			fetch = &models.Fetch{
				ID:             r.FetchID,
				InitialAddress: r.InitialAddress,
				Quantity:       r.Quantity,
				FunctionCode:   r.FunctionCode,
				Tags:           make([]*models.Tag, 0),
			}
			fetchMap[fetchKey] = fetch

			dev.Fetches = append(dev.Fetches, fetch)
		}

		tagKey := fetchKey + ":" + r.TagID
		tag, tExists := tagMap[tagKey]
		if !tExists {
			tag = &models.Tag{
				ID:             r.TagID,
				Name:           r.TagName,
				DataType:       r.DataType,
				Operation:      r.Operation,
				OperationValue: r.OperationValue,
				Positions:      make([]int, 0),
			}
			tagMap[tagKey] = tag

			fetch.Tags = append(fetch.Tags, tag)
		}

		tag.Positions = append(tag.Positions, r.Position)
	}

	result := make([]models.Device, 0, len(deviceOrder))
	for _, id := range deviceOrder {
		result = append(result, *deviceMap[id])
	}

	logger.Info("LoadDevices executed successfully", zap.String("journey", "database"))

	return result, nil
}
