//go:build integration

package mqtt

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"modbus-driver/internal/config"
	"modbus-driver/internal/models"

	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Helpers ─────────────────────────────────────────────────────────────────

func loadTestEnv(t *testing.T) *config.MQTT {
	t.Helper()

	// Try loading .env.test from project root
	_ = godotenv.Load("../../.env.test")

	broker := os.Getenv("TEST_MQTT_BROKER")
	if broker == "" {
		t.Skip("TEST_MQTT_BROKER not set — skipping integration test")
	}

	return &config.MQTT{
		Broker:     broker,
		ClientID:   os.Getenv("TEST_MQTT_CLIENT_ID"),
		Username:   os.Getenv("TEST_MQTT_USERNAME"),
		Password:   os.Getenv("TEST_MQTT_PASSWORD"),
		TenantCode: os.Getenv("TEST_TENANT_CODE"),
	}
}

// spySubscriber connects to the broker and subscribes to the given topic.
// It returns a channel that will receive each payload received on that topic.
func spySubscriber(t *testing.T, cfg *config.MQTT, topic string) <-chan string {
	t.Helper()

	payloads := make(chan string, 100)

	opts := paho.NewClientOptions().
		AddBroker(cfg.Broker).
		SetClientID(cfg.ClientID + "-spy-" + fmt.Sprintf("%d", time.Now().UnixNano())).
		SetCleanSession(true).
		SetAutoReconnect(false).
		SetConnectTimeout(5 * time.Second)

	if cfg.Username != "" {
		opts.SetUsername(cfg.Username).SetPassword(cfg.Password)
	}

	spy := paho.NewClient(opts)
	tok := spy.Connect()
	if !tok.WaitTimeout(5 * time.Second) {
		t.Skip("Could not connect spy to broker — skipping integration test")
	}
	if tok.Error() != nil {
		t.Skip("Spy connect error — skipping integration test: " + tok.Error().Error())
	}

	subTok := spy.Subscribe(topic, 0, func(_ paho.Client, msg paho.Message) {
		payloads <- string(msg.Payload())
	})
	require.True(t, subTok.WaitTimeout(5*time.Second), "subscribe timed out")
	require.NoError(t, subTok.Error())

	t.Cleanup(func() {
		spy.Unsubscribe(topic)
		spy.Disconnect(250)
		close(payloads)
	})

	// Small delay so the subscription is established on the broker before publishing
	time.Sleep(500 * time.Millisecond)

	return payloads
}

// ─── Integration Tests ──────────────────────────────────────────────────────

func TestIntegration_PublishAndReceive(t *testing.T) {
	cfg := loadTestEnv(t)

	// 1. Create spy subscriber on the topic we'll publish to
	topic := fmt.Sprintf("formatted/%s/int-test-device/int-test-tag", cfg.TenantCode)
	payloads := spySubscriber(t, cfg, topic)

	// 2. Create our real Publisher
	pub, err := NewPublisher(cfg)
	require.NoError(t, err, "NewPublisher should connect to the test broker")
	t.Cleanup(func() { pub.client.Disconnect(250) })

	// 3. Publish a payload
	payload := models.MQTTPayload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Val:       "42.5",
	}
	jsonBytes, _ := json.Marshal(payload)
	pub.Publish("int-test-device", "int-test-tag", string(jsonBytes))

	// 4. Wait for the spy to receive it
	select {
	case received := <-payloads:
		var got models.MQTTPayload
		err := json.Unmarshal([]byte(received), &got)
		require.NoError(t, err)
		assert.Equal(t, "42.5", got.Val)
		assert.NotEmpty(t, got.Timestamp)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for MQTT message from spy subscriber")
	}
}

func TestIntegration_PublishHealthAndReceive(t *testing.T) {
	cfg := loadTestEnv(t)

	topic := fmt.Sprintf("health/%s/int-test-device", cfg.TenantCode)
	payloads := spySubscriber(t, cfg, topic)

	pub, err := NewPublisher(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { pub.client.Disconnect(250) })

	hp := models.HealthPayload{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Status:    "online",
	}
	jsonBytes, _ := json.Marshal(hp)
	pub.PublishHealth("int-test-device", string(jsonBytes))

	select {
	case received := <-payloads:
		var got models.HealthPayload
		err := json.Unmarshal([]byte(received), &got)
		require.NoError(t, err)
		assert.Equal(t, "online", got.Status)
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for health MQTT message")
	}
}

func TestIntegration_RetainedMessage(t *testing.T) {
	cfg := loadTestEnv(t)

	// 1. Publish first (with retain=true, which is what our Publisher uses)
	pub, err := NewPublisher(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { pub.client.Disconnect(250) })

	topic := fmt.Sprintf("formatted/%s/retain-test/tag-1", cfg.TenantCode)
	payload := `{"TS":"2026-01-01T00:00:00Z","Val":"retained-value"}`
	pub.Publish("retain-test", "tag-1", payload)

	// Wait for the publish to propagate
	time.Sleep(500 * time.Millisecond)

	// 2. Now subscribe AFTER publishing — retained message should arrive immediately
	payloads := spySubscriber(t, cfg, topic)

	select {
	case received := <-payloads:
		assert.Contains(t, received, "retained-value",
			"should receive the retained message even after subscribing late")
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for retained MQTT message")
	}
}

func TestIntegration_MultipleTopics_Concurrent(t *testing.T) {
	cfg := loadTestEnv(t)

	pub, err := NewPublisher(cfg)
	require.NoError(t, err)
	t.Cleanup(func() { pub.client.Disconnect(250) })

	const numDevices = 10
	spies := make([]<-chan string, numDevices)

	// Subscribe to 10 different topics
	for i := 0; i < numDevices; i++ {
		deviceID := fmt.Sprintf("concurrent-dev-%d", i)
		topic := fmt.Sprintf("formatted/%s/%s/tag-0", cfg.TenantCode, deviceID)
		spies[i] = spySubscriber(t, cfg, topic)
	}

	// Publish to all 10 concurrently
	var wg sync.WaitGroup
	for i := 0; i < numDevices; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			deviceID := fmt.Sprintf("concurrent-dev-%d", idx)
			payload := fmt.Sprintf(`{"TS":"2026-01-01T00:00:00Z","Val":"%d"}`, idx)
			pub.Publish(deviceID, "tag-0", payload)
		}(i)
	}
	wg.Wait()

	// Assert all 10 spies received their message
	for i := 0; i < numDevices; i++ {
		select {
		case received := <-spies[i]:
			expected := fmt.Sprintf(`"%d"`, i)
			assert.Contains(t, received, expected)
		case <-time.After(5 * time.Second):
			t.Fatalf("device %d: timed out waiting for message", i)
		}
	}
}
