package config

import (
	"modbus-driver/internal/config/logger"
	"os"
	"strconv"
	"time"

	"go.uber.org/zap"
)

var (
	POLL_INTERVAL     = "POLL_INTERVAL"
	MODBUS_TIMEOUT    = "MODBUS_TIMEOUT"
	DEVICE_GROUP_SIZE = "DEVICE_GROUP_SIZE"
)

type Modbus struct {
	Timeout         time.Duration
	PollInterval    time.Duration
	DeviceGroupSize int
}

func NewModbusConfig() *Modbus {
	logger.Info("Init NewModbusConfig", zap.String("journey", "modbus"))

	modbusTimeout, err := time.ParseDuration(os.Getenv(MODBUS_TIMEOUT))
	if err != nil {
		logger.Error("ParseDuration func returned an error", err, zap.String("journey", "modbus"))
	}

	poolInterval, err := time.ParseDuration(os.Getenv(POLL_INTERVAL))
	if err != nil {
		logger.Error("ParseDuration func returned an error", err, zap.String("journey", "modbus"))
	}

	deviceGroupSize, err := strconv.Atoi(os.Getenv(DEVICE_GROUP_SIZE))
	if err != nil {
		logger.Error("Atoi func returned an error", err, zap.String("journey", "modbus"))
	}

	mb := &Modbus{
		Timeout:         modbusTimeout,
		PollInterval:    poolInterval,
		DeviceGroupSize: deviceGroupSize,
	}

	logger.Info("NewModbusConfig executed successfully", zap.String("journey", "modbus"))

	return mb
}
