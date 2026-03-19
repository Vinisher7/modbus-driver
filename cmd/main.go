package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"modbus-driver/internal/config"
	"modbus-driver/internal/database"
	modbuspkg "modbus-driver/internal/modbus"
	mqttpkg "modbus-driver/internal/mqtt"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// ── Banco de dados ────────────────────────────────────────────────
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()
	log.Println("[DB] connected to SQL Server")

	// ── Carrega metadados no startup ──────────────────────────────────
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	devices, err := database.LoadDevices(ctx, db)
	if err != nil {
		log.Fatalf("load devices: %v", err)
	}
	log.Printf("[DB] loaded %d modbus devices", len(devices))

	if len(devices) == 0 {
		log.Fatal("no modbus devices found — check the database and protocol column")
	}

	// ── MQTT Publisher ────────────────────────────────────────────────
	pub, err := mqttpkg.NewPublisher(cfg)
	if err != nil {
		log.Fatalf("mqtt: %v", err)
	}

	// ── Driver ───────────────────────────────────────────────────────
	driver := modbuspkg.NewDriver(cfg, pub)
	go driver.RunPollLoop(ctx, devices)

	// ── Graceful shutdown ─────────────────────────────────────────────
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	cancel()
}
