package config

import (
	"modbus-driver/internal/config/logger"
	"os"

	"go.uber.org/zap"
)

var (
	MQTT_BROKER    = "MQTT_BROKER"
	MQTT_CLIENT_ID = "MQTT_CLIENT_ID"
	MQTT_USERNAME  = "MQTT_USERNAME"
	MQTT_PASSWORD  = "MQTT_PASSWORD"
	TENANT_CODE    = "TENANT_CODE"
)

type MQTT struct {
	Broker     string
	ClientID   string
	Username   string
	Password   string
	TenantCode string
}

func NewMqttConfig() *MQTT {
	logger.Info("Init NewMqttConfig", zap.String("journey", "mqtt"))

	mqtt := &MQTT{
		Broker:     os.Getenv(MQTT_BROKER),
		ClientID:   os.Getenv(MQTT_CLIENT_ID),
		Username:   os.Getenv(MQTT_USERNAME),
		Password:   os.Getenv(MQTT_PASSWORD),
		TenantCode: os.Getenv(TENANT_CODE),
	}

	logger.Info("NewMqttConfig executed successfully", zap.String("journey", "mqtt"))

	return mqtt
}
