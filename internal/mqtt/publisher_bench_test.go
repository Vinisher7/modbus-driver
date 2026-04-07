//go:build integration

package mqtt

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"modbus-driver/internal/config"
	"modbus-driver/internal/models"

	"github.com/joho/godotenv"
)

// loadBenchEnv loads the .env.test and creates an MQTT config.
// If the broker is not configured, the benchmark will skip.
func loadBenchEnv(b *testing.B) *config.MQTT {
	b.Helper()

	_ = godotenv.Load("../../.env.test")

	broker := os.Getenv("TEST_MQTT_BROKER")
	if broker == "" {
		b.Skip("TEST_MQTT_BROKER not set — skipping benchmark")
	}

	return &config.MQTT{
		Broker:     broker,
		ClientID:   os.Getenv("TEST_MQTT_CLIENT_ID"),
		Username:   os.Getenv("TEST_MQTT_USERNAME"),
		Password:   os.Getenv("TEST_MQTT_PASSWORD"),
		TenantCode: os.Getenv("TEST_TENANT_CODE"),
	}
}

// ─── Benchmark: Raw Publish Throughput ──────────────────────────────────────
// Measures how many MQTT publishes per second our Publisher can sustain
// under the current synchronous tok.Wait() implementation.
//
// Run with:
//   go test -tags=integration -bench=BenchmarkPublish -benchmem -benchtime=5s ./internal/mqtt/

func BenchmarkPublish_SingleTopic(b *testing.B) {
	cfg := loadBenchEnv(b)

	pub, err := NewPublisher(cfg)
	if err != nil {
		b.Fatalf("NewPublisher: %v", err)
	}
	defer pub.client.Disconnect(250)

	payload := models.MQTTPayload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Val:       "123.456",
	}
	jsonBytes, _ := json.Marshal(payload)
	msg := string(jsonBytes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pub.Publish("bench-device", "bench-tag", msg)
	}
}

func BenchmarkPublish_MultipleTopic(b *testing.B) {
	cfg := loadBenchEnv(b)

	pub, err := NewPublisher(cfg)
	if err != nil {
		b.Fatalf("NewPublisher: %v", err)
	}
	defer pub.client.Disconnect(250)

	payload := models.MQTTPayload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Val:       "789.012",
	}
	jsonBytes, _ := json.Marshal(payload)
	msg := string(jsonBytes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deviceID := fmt.Sprintf("bench-dev-%d", i%50)
		tagID := fmt.Sprintf("bench-tag-%d", i%100)
		pub.Publish(deviceID, tagID, msg)
	}
}

func BenchmarkPublishHealth(b *testing.B) {
	cfg := loadBenchEnv(b)

	pub, err := NewPublisher(cfg)
	if err != nil {
		b.Fatalf("NewPublisher: %v", err)
	}
	defer pub.client.Disconnect(250)

	hp := models.HealthPayload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Status:    "online",
	}
	jsonBytes, _ := json.Marshal(hp)
	msg := string(jsonBytes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		deviceID := fmt.Sprintf("bench-health-%d", i%50)
		pub.PublishHealth(deviceID, msg)
	}
}

// ─── Benchmark: Parallel Publish (simulating concurrent goroutines) ─────────
// This simulates the real-world scenario where multiple goroutines from
// pollAllGroups are publishing concurrently through a shared Publisher.
//
// Run with:
//   go test -tags=integration -bench=BenchmarkPublish_Parallel -benchmem -benchtime=5s ./internal/mqtt/

func BenchmarkPublish_Parallel(b *testing.B) {
	cfg := loadBenchEnv(b)

	pub, err := NewPublisher(cfg)
	if err != nil {
		b.Fatalf("NewPublisher: %v", err)
	}
	defer pub.client.Disconnect(250)

	payload := models.MQTTPayload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Val:       "parallel-42",
	}
	jsonBytes, _ := json.Marshal(payload)
	msg := string(jsonBytes)

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			deviceID := fmt.Sprintf("par-dev-%d", i%50)
			tagID := fmt.Sprintf("par-tag-%d", i%100)
			pub.Publish(deviceID, tagID, msg)
			i++
		}
	})
}
