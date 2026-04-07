package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewMqttConfig_AllEnvsSet(t *testing.T) {
	t.Setenv("MQTT_BROKER", "tcp://broker.local:1883")
	t.Setenv("MQTT_CLIENT_ID", "driver-01")
	t.Setenv("MQTT_USERNAME", "admin")
	t.Setenv("MQTT_PASSWORD", "secret123")
	t.Setenv("TENANT_CODE", "tenant-abc")

	cfg := NewMqttConfig()

	assert.Equal(t, "tcp://broker.local:1883", cfg.Broker)
	assert.Equal(t, "driver-01", cfg.ClientID)
	assert.Equal(t, "admin", cfg.Username)
	assert.Equal(t, "secret123", cfg.Password)
	assert.Equal(t, "tenant-abc", cfg.TenantCode)
}

func TestNewMqttConfig_EmptyEnvs(t *testing.T) {
	t.Setenv("MQTT_BROKER", "")
	t.Setenv("MQTT_CLIENT_ID", "")
	t.Setenv("MQTT_USERNAME", "")
	t.Setenv("MQTT_PASSWORD", "")
	t.Setenv("TENANT_CODE", "")

	cfg := NewMqttConfig()

	assert.Empty(t, cfg.Broker)
	assert.Empty(t, cfg.ClientID)
	assert.Empty(t, cfg.Username)
	assert.Empty(t, cfg.Password)
	assert.Empty(t, cfg.TenantCode)
}

func TestNewMqttConfig_PartialEnvs(t *testing.T) {
	t.Setenv("MQTT_BROKER", "tcp://192.168.1.10:1883")
	t.Setenv("MQTT_CLIENT_ID", "")
	t.Setenv("MQTT_USERNAME", "user")
	t.Setenv("MQTT_PASSWORD", "")
	t.Setenv("TENANT_CODE", "factory-01")

	cfg := NewMqttConfig()

	assert.Equal(t, "tcp://192.168.1.10:1883", cfg.Broker)
	assert.Empty(t, cfg.ClientID)
	assert.Equal(t, "user", cfg.Username)
	assert.Empty(t, cfg.Password)
	assert.Equal(t, "factory-01", cfg.TenantCode)
}
