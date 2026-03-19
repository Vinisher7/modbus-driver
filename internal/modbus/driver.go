package modbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"sync"
	"time"

	"modbus-driver/internal/config"
	"modbus-driver/internal/models"
	mqttpkg "modbus-driver/internal/mqtt"
)

type Driver struct {
	cfg       *config.Config
	publisher *mqttpkg.Publisher
}

func NewDriver(cfg *config.Config, pub *mqttpkg.Publisher) *Driver {
	return &Driver{cfg: cfg, publisher: pub}
}

// RunPollLoop executa o ciclo de polling continuamente
func (d *Driver) RunPollLoop(ctx context.Context, devices []models.Device) {
	ticker := time.NewTicker(d.cfg.PollInterval)
	defer ticker.Stop()

	log.Printf("[Driver] polling %d devices, group size %d, interval %s",
		len(devices), d.cfg.DeviceGroupSize, d.cfg.PollInterval)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			d.pollAllGroups(ctx, devices)
		}
	}
}

// pollAllGroups divide os devices em grupos de DeviceGroupSize e processa em paralelo
func (d *Driver) pollAllGroups(ctx context.Context, devices []models.Device) {
	groups := chunkDevices(devices, d.cfg.DeviceGroupSize)
	log.Printf("[Driver] starting %d group(s)", len(groups))

	var wg sync.WaitGroup
	for i, group := range groups {
		wg.Add(1)
		go func(groupIdx int, grp []models.Device) {
			defer wg.Done()
			d.processGroup(ctx, groupIdx, grp)
		}(i, group)
	}
	wg.Wait()
	log.Println("[Driver] poll cycle complete")
}

func (d *Driver) processGroup(ctx context.Context, groupIdx int, devices []models.Device) {
	var wg sync.WaitGroup
	for _, dev := range devices {
		wg.Add(1)
		go func(dev models.Device) {
			defer wg.Done()
			if err := d.pollDevice(ctx, dev); err != nil {
				log.Printf("[Driver] group=%d device=%s err=%v", groupIdx, dev.DeviceID, err)
			}
		}(dev)
	}
	wg.Wait()
}

func (d *Driver) pollDevice(ctx context.Context, dev models.Device) error {
	cli, err := NewClient(dev.Host, dev.Port, dev.UnitID, d.cfg.ModbusTimeout)
	if err != nil {
		return err
	}
	defer cli.Close()

	for _, fetch := range dev.Fetches {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		regs, err := cli.ReadRegisters(fetch.FunctionCode, fetch.InitialAddress, fetch.Quantity)
		if err != nil {
			log.Printf("[Driver] device=%s fetch=%s read error: %v", dev.DeviceID, fetch.ID, err)
			continue
		}

		for _, tag := range fetch.Tags {
			// valida se todas as positions estão dentro do range
			outOfRange := false
			for _, pos := range tag.Positions {
				if pos >= len(regs) {
					log.Printf("[Driver] tag=%s position=%d out of range (len=%d)", tag.Name, pos, len(regs))
					outOfRange = true
					break
				}
			}
			if outOfRange {
				continue
			}

			// combina words e converte conforme data_type
			val, err := combineRegisters(regs, tag.Positions, tag.DataType)
			if err != nil {
				log.Printf("[Driver] tag=%s combine error: %v", tag.Name, err)
				continue
			}

			finalVal := applyOperation(val, tag.Operation, tag.OperationValue)

			payload := models.MQTTPayload{
				Timestamp: time.Now().UTC().Format(time.RFC3339),
				Val:       finalVal,
			}
			jsonBytes, _ := json.Marshal(payload)
			d.publisher.Publish(dev.DeviceID, tag.ID, string(jsonBytes))
		}
	}
	return nil
}

// applyOperation aplica operação opcional no valor do registro
func applyOperation(val float64, op, opVal *string) string {
	if op == nil || opVal == nil {
		return fmt.Sprintf("%v", val)
	}
	var operand float64
	fmt.Sscanf(*opVal, "%f", &operand)

	switch *op {
	case "*":
		val *= operand
	case "/":
		if operand != 0 {
			val /= operand
		}
	case "+":
		val += operand
	case "-":
		val -= operand
	}
	return fmt.Sprintf("%v", val)
}

func chunkDevices(devices []models.Device, size int) [][]models.Device {
	var chunks [][]models.Device
	for size < len(devices) {
		devices, chunks = devices[size:], append(chunks, devices[:size])
	}
	return append(chunks, devices)
}

func combineRegisters(regs []uint16, positions []int, dataType string) (float64, error) {
	wordCount := len(positions)

	if dataType == "auto" {
		switch wordCount {
		case 1:
			dataType = "uint16"
		case 2:
			dataType = "float32"
		default:
			return 0, fmt.Errorf("auto não suporta %d words", wordCount)
		}
	}

	switch dataType {
	case "uint16":
		return float64(regs[positions[0]]), nil

	case "int16":
		return float64(int16(regs[positions[0]])), nil

	case "float32":
		raw := uint32(regs[positions[0]])<<16 | uint32(regs[positions[1]])
		return float64(math.Float32frombits(raw)), nil

	case "float32_swapped":
		raw := uint32(regs[positions[1]])<<16 | uint32(regs[positions[0]])
		return float64(math.Float32frombits(raw)), nil

	case "int32":
		raw := int32(regs[positions[0]])<<16 | int32(regs[positions[1]])
		return float64(raw), nil

	case "uint32":
		raw := uint32(regs[positions[0]])<<16 | uint32(regs[positions[1]])
		return float64(raw), nil
	}

	return 0, fmt.Errorf("data_type '%s' desconhecido", dataType)
}
