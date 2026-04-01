package modbus

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"modbus-driver/internal/config"
	"modbus-driver/internal/config/logger"
	"modbus-driver/internal/models"
	mqttpkg "modbus-driver/internal/mqtt"

	"go.uber.org/zap"
)

type Driver struct {
	publisher *mqttpkg.Publisher
}

func NewDriver(pub *mqttpkg.Publisher) *Driver {
	logger.Info("Init NewDriver", zap.String("journey", "modbus"))

	driver := &Driver{
		publisher: pub,
	}

	logger.Info("NewDriver executed successfully", zap.String("journey", "modbus"))
	return driver
}

func (d *Driver) RunPollLoop(ctx context.Context, devices []models.Device, mb *config.Modbus) {
	logger.Info("Init RunPollLoop", zap.String("journey", "modbus"))

	ticker := time.NewTicker(mb.PollInterval)
	defer ticker.Stop()

	message := fmt.Sprintf("Polling devices=%d groups=%d interval=%s", len(devices), mb.DeviceGroupSize, mb.PollInterval)
	logger.Info(message, zap.String("journey", "modbus"))

	for {
		select {
		case <-ctx.Done():
			logger.Error("RunPollLoop func returned an error", ctx.Err(), zap.String("journey", "modbus"))
			return
		case <-ticker.C:
			logger.Info("RunPollLoop triggered successfully", zap.String("journey", "modbus"))
			d.pollAllGroups(ctx, devices, mb)
		}
	}
}

func (d *Driver) pollAllGroups(ctx context.Context, devices []models.Device, mb *config.Modbus) {
	logger.Info("Init pollAllGroups", zap.String("journey", "modbus"))

	groups := chunkDevices(devices, mb.DeviceGroupSize)
	message := fmt.Sprintf("starting %d group(s)", len(groups))
	logger.Info(message, zap.String("journey", "modbus"))

	var wg sync.WaitGroup
	for i, group := range groups {
		wg.Add(1)
		go func(groupIdx int, grp []models.Device) {
			defer wg.Done()
			d.processGroup(ctx, groupIdx, grp, mb)
		}(i, group)
	}
	wg.Wait()
	logger.Info("pollAllGroups executed successfully", zap.String("journey", "modbus"))
}

func (d *Driver) processGroup(ctx context.Context, groupIdx int, devices []models.Device, mb *config.Modbus) {
	logger.Info("Init processGroup", zap.String("journey", "modbus"))

	var wg sync.WaitGroup
	for _, dev := range devices {
		wg.Add(1)
		go func(dev models.Device) {
			defer wg.Done()
			if err := d.pollDevice(ctx, dev, mb); err != nil {
				logger.Error("pollDevice func returned an error", err, zap.String("journey", "modbus"))
			}
		}(dev)
	}
	wg.Wait()
	logger.Info("processGroup executed successfully", zap.String("journey", "modbus"))
}

func (d *Driver) pollDevice(ctx context.Context, dev models.Device, mb *config.Modbus) error {
	logger.Info("Init pollDevice", zap.String("journey", "modbus"))

	cli, err := NewClient(dev.Host, dev.Port, dev.UnitID, mb.Timeout)
	if err != nil {
		logger.Error("NewClient func returned an error", err, zap.String("journey", "modbus"))
		return err
	}
	defer cli.Close()

	for _, fetch := range dev.Fetches {
		select {
		case <-ctx.Done():
			logger.Error("Context was canceled", ctx.Err(), zap.String("journey", "modbus"))
			return ctx.Err()
		default:
		}

		regs, err := cli.ReadRegisters(fetch.FunctionCode, fetch.InitialAddress, fetch.Quantity)
		if err != nil {
			message := fmt.Sprintf("device=%s fetch=%s returned an error", dev.DeviceID, fetch.ID)
			logger.Error(message, err, zap.String("journey", "modbus"))
			continue
		}

		for _, tag := range fetch.Tags {
			outOfRange := false
			for _, pos := range tag.Positions {
				if pos >= len(regs) {
					message := fmt.Sprintf("tag=%s position=%d out of range (len=%d)", tag.Name, pos, len(regs))
					logger.Error(message, err, zap.String("journey", "modbus"))

					outOfRange = true
					break
				}
			}
			if outOfRange {
				continue
			}

			val, err := combineRegisters(regs, tag.Positions, tag.DataType)
			if err != nil {
				message := fmt.Sprintf("tag=%s combine error", tag.Name)
				logger.Error(message, err, zap.String("journey", "modbus"))

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

	logger.Info("pollDevice executed successfully", zap.String("journey", "modbus"))
	return nil
}

func applyOperation(val float64, op, opVal *string) string {
	logger.Info("Init applyOperation", zap.String("journey", "modbus"))

	if op == nil || opVal == nil {
		logger.Info("applyOperation executed successfully (no op)", zap.String("journey", "modbus"))
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

	logger.Info("applyOperation executed successfully", zap.String("journey", "modbus"))
	return fmt.Sprintf("%v", val)
}

func chunkDevices(devices []models.Device, size int) [][]models.Device {
	logger.Info("Init chunkDevices", zap.String("journey", "modbus"))

	var chunks [][]models.Device
	for size < len(devices) {
		devices, chunks = devices[size:], append(chunks, devices[:size])
	}

	logger.Info("chunkDevices executed successfully", zap.String("journey", "modbus"))
	return append(chunks, devices)
}

func combineRegisters(regs []uint16, positions []int, dataType string) (float64, error) {
	logger.Info("Init combineRegisters", zap.String("journey", "modbus"))

	wordCount := len(positions)

	if dataType == "auto" {
		switch wordCount {
		case 1:
			dataType = "uint16"
		case 2:
			dataType = "float32"
		default:
			err := fmt.Errorf("auto não suporta %d words", wordCount)
			logger.Error("combineRegisters auto returned an error", err, zap.String("journey", "modbus"))
			return 0, err
		}
	}

	var result float64
	var err error

	switch dataType {
	case "uint16":
		result = float64(regs[positions[0]])
	case "int16":
		result = float64(int16(regs[positions[0]]))
	case "float32":
		raw := uint32(regs[positions[0]])<<16 | uint32(regs[positions[1]])
		result = float64(math.Float32frombits(raw))
	case "float32_swapped":
		raw := uint32(regs[positions[1]])<<16 | uint32(regs[positions[0]])
		result = float64(math.Float32frombits(raw))
	case "int32":
		raw := int32(regs[positions[0]])<<16 | int32(regs[positions[1]])
		result = float64(raw)
	case "uint32":
		raw := uint32(regs[positions[0]])<<16 | uint32(regs[positions[1]])
		result = float64(raw)
	default:
		err = fmt.Errorf("data_type '%s' desconhecido", dataType)
	}

	if err != nil {
		logger.Error("combineRegisters switch returned an error", err, zap.String("journey", "modbus"))
		return 0, err
	}

	logger.Info("combineRegisters executed successfully", zap.String("journey", "modbus"))
	return result, nil
}
