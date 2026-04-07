package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewModbusConfig_AllEnvsSet(t *testing.T) {
	t.Setenv("MODBUS_TIMEOUT", "3s")
	t.Setenv("POLL_INTERVAL", "5s")
	t.Setenv("DEVICE_GROUP_SIZE", "10")

	cfg := NewModbusConfig()

	assert.Equal(t, 3*time.Second, cfg.Timeout)
	assert.Equal(t, 5*time.Second, cfg.PollInterval)
	assert.Equal(t, 10, cfg.DeviceGroupSize)
}

func TestNewModbusConfig_MillisecondPrecision(t *testing.T) {
	t.Setenv("MODBUS_TIMEOUT", "500ms")
	t.Setenv("POLL_INTERVAL", "1500ms")
	t.Setenv("DEVICE_GROUP_SIZE", "5")

	cfg := NewModbusConfig()

	assert.Equal(t, 500*time.Millisecond, cfg.Timeout)
	assert.Equal(t, 1500*time.Millisecond, cfg.PollInterval)
	assert.Equal(t, 5, cfg.DeviceGroupSize)
}

func TestNewModbusConfig_InvalidEnvs_FallsBackToZero(t *testing.T) {
	t.Setenv("MODBUS_TIMEOUT", "invalid")
	t.Setenv("POLL_INTERVAL", "not-a-duration")
	t.Setenv("DEVICE_GROUP_SIZE", "abc")

	cfg := NewModbusConfig()

	assert.Equal(t, time.Duration(0), cfg.Timeout)
	assert.Equal(t, time.Duration(0), cfg.PollInterval)
	assert.Equal(t, 0, cfg.DeviceGroupSize)
}

func TestNewModbusConfig_EmptyEnvs_FallsBackToZero(t *testing.T) {
	t.Setenv("MODBUS_TIMEOUT", "")
	t.Setenv("POLL_INTERVAL", "")
	t.Setenv("DEVICE_GROUP_SIZE", "")

	cfg := NewModbusConfig()

	assert.Equal(t, time.Duration(0), cfg.Timeout)
	assert.Equal(t, time.Duration(0), cfg.PollInterval)
	assert.Equal(t, 0, cfg.DeviceGroupSize)
}
