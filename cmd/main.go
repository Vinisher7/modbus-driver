package main

import (
	"context"
	"modbus-driver/internal/config"
	"modbus-driver/internal/config/database"
	"modbus-driver/internal/config/logger"
	"modbus-driver/internal/modbus"
	"modbus-driver/internal/mqtt"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/zap"
)

func main() {
	logger.Info("Starting application...")

	// // ── Banco de dados ────────────────────────────────────────────────
	ctx := context.Background()

	db, err := database.NewPostgresClient(ctx)
	if err != nil {
		logger.Error("NewPostgresClient func returned an error", err, zap.String("journey", "database"))
		panic(err)
	}
	defer db.Close(ctx)

	err = database.TestPostgresConnection(ctx, db)
	if err != nil {
		logger.Error("TestPostgresConnection func returned an error", err, zap.String("journey", "database"))
		panic(err)
	}

	devices, err := database.LoadDevices(ctx, db)
	if err != nil {
		logger.Error("LoadDevices func returned an error", err, zap.String("journey", "database"))
		panic(err)
	}

	// ── MQTT Publisher ────────────────────────────────────────────────
	pub, err := mqtt.NewPublisher(config.NewMqttConfig())
	if err != nil {
		logger.Error("NewPublisher func returned an error", err, zap.String("journey", "publisher"))
		panic(err)
	}

	// ── Driver ───────────────────────────────────────────────────────
	driver := modbus.NewDriver(pub)
	go driver.RunPollLoop(ctx, devices, config.NewModbusConfig())

	// ── Graceful shutdown ─────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down application...")

}
