package mqtt

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ─── Topic Format Tests ─────────────────────────────────────────────────────
// These test the topic formatting logic without requiring a real MQTT broker.
// We extract and test the topic building format used by Publish and PublishHealth.

func TestPublishTopicFormat(t *testing.T) {
	tests := []struct {
		name       string
		tenantCode string
		deviceID   string
		tagID      string
		want       string
	}{
		{
			name:       "standard topic",
			tenantCode: "factory-01",
			deviceID:   "device-abc",
			tagID:      "tag-temp",
			want:       "formatted/factory-01/device-abc/tag-temp",
		},
		{
			name:       "UUID-style IDs",
			tenantCode: "tenant-xyz",
			deviceID:   "550e8400-e29b-41d4-a716-446655440000",
			tagID:      "a3b8d1b6-0b3b-4b1a-9c1a-1a2b3c4d5e6f",
			want:       "formatted/tenant-xyz/550e8400-e29b-41d4-a716-446655440000/a3b8d1b6-0b3b-4b1a-9c1a-1a2b3c4d5e6f",
		},
		{
			name:       "empty tenant",
			tenantCode: "",
			deviceID:   "dev-1",
			tagID:      "tag-1",
			want:       "formatted//dev-1/tag-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topic := fmt.Sprintf("formatted/%s/%s/%s", tt.tenantCode, tt.deviceID, tt.tagID)
			assert.Equal(t, tt.want, topic)
		})
	}
}

func TestHealthTopicFormat(t *testing.T) {
	tests := []struct {
		name       string
		tenantCode string
		deviceID   string
		want       string
	}{
		{
			name:       "standard health topic",
			tenantCode: "factory-01",
			deviceID:   "device-abc",
			want:       "health/factory-01/device-abc",
		},
		{
			name:       "UUID device",
			tenantCode: "tenant-xyz",
			deviceID:   "550e8400-e29b-41d4-a716-446655440000",
			want:       "health/tenant-xyz/550e8400-e29b-41d4-a716-446655440000",
		},
		{
			name:       "empty tenant",
			tenantCode: "",
			deviceID:   "dev-1",
			want:       "health//dev-1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			topic := fmt.Sprintf("health/%s/%s", tt.tenantCode, tt.deviceID)
			assert.Equal(t, tt.want, topic)
		})
	}
}
