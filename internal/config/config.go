package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DBServer        string
	DBPort          int
	DBUser          string
	DBPassword      string
	DBName          string
	MQTTBroker      string
	MQTTClientID    string
	MQTTUsername    string
	MQTTPassword    string
	ModbusTimeout   time.Duration
	PollInterval    time.Duration
	DeviceGroupSize int
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	dbPort, _ := strconv.Atoi(getEnv("DB_PORT", "1433"))
	modbusTimeout, _ := time.ParseDuration(getEnv("MODBUS_TIMEOUT", "5s"))
	pollInterval, _ := time.ParseDuration(getEnv("POLL_INTERVAL", "10s"))
	groupSize, _ := strconv.Atoi(getEnv("DEVICE_GROUP_SIZE", "5"))

	return &Config{
		DBServer:        getEnv("DB_SERVER", "localhost"),
		DBPort:          dbPort,
		DBUser:          getEnv("DB_USER", "sa"),
		DBPassword:      getEnv("DB_PASSWORD", ""),
		DBName:          getEnv("DB_NAME", ""),
		MQTTBroker:      getEnv("MQTT_BROKER", "tcp://localhost:1883"),
		MQTTClientID:    getEnv("MQTT_CLIENT_ID", "modbus-driver"),
		MQTTUsername:    getEnv("MQTT_USERNAME", ""),
		MQTTPassword:    getEnv("MQTT_PASSWORD", ""),
		ModbusTimeout:   modbusTimeout,
		PollInterval:    pollInterval,
		DeviceGroupSize: groupSize,
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
