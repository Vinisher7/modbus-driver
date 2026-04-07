package modbus

import (
	"encoding/json"
	"math"
	"sync"
	"testing"
	"time"

	"modbus-driver/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── Mock Publisher ──────────────────────────────────────────────────────────

type publishCall struct {
	DeviceID string
	TagID    string
	Payload  string
}

type healthCall struct {
	DeviceID string
	Payload  string
}

type MockPublisher struct {
	mu          sync.Mutex
	publishes   []publishCall
	healthCalls []healthCall
}

func (m *MockPublisher) RecordPublish(deviceID, tagID, payload string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishes = append(m.publishes, publishCall{deviceID, tagID, payload})
}

func (m *MockPublisher) PublishHealth(deviceID, payload string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.healthCalls = append(m.healthCalls, healthCall{deviceID, payload})
}

func (m *MockPublisher) GetHealthCalls() []healthCall {
	m.mu.Lock()
	defer m.mu.Unlock()
	dst := make([]healthCall, len(m.healthCalls))
	copy(dst, m.healthCalls)
	return dst
}

// ─── chunkDevices ────────────────────────────────────────────────────────────

func TestChunkDevices_EvenSplit(t *testing.T) {
	devices := make([]models.Device, 6)
	for i := range devices {
		devices[i] = models.Device{DeviceID: string(rune('A' + i))}
	}

	chunks := chunkDevices(devices, 2)

	assert.Len(t, chunks, 3)
	for _, chunk := range chunks {
		assert.Len(t, chunk, 2)
	}
}

func TestChunkDevices_UnevenSplit(t *testing.T) {
	devices := make([]models.Device, 5)
	for i := range devices {
		devices[i] = models.Device{DeviceID: string(rune('A' + i))}
	}

	chunks := chunkDevices(devices, 3)

	assert.Len(t, chunks, 2)
	assert.Len(t, chunks[0], 3)
	assert.Len(t, chunks[1], 2)
}

func TestChunkDevices_SizeGreaterThanDevices(t *testing.T) {
	devices := []models.Device{{DeviceID: "A"}, {DeviceID: "B"}}

	chunks := chunkDevices(devices, 10)

	assert.Len(t, chunks, 1)
	assert.Len(t, chunks[0], 2)
}

func TestChunkDevices_SingleDevice(t *testing.T) {
	devices := []models.Device{{DeviceID: "only"}}

	chunks := chunkDevices(devices, 1)

	assert.Len(t, chunks, 1)
	assert.Equal(t, "only", chunks[0][0].DeviceID)
}

func TestChunkDevices_EmptySlice(t *testing.T) {
	chunks := chunkDevices([]models.Device{}, 5)

	assert.Len(t, chunks, 1)
	assert.Empty(t, chunks[0])
}

// ─── combineRegisters ────────────────────────────────────────────────────────

func TestCombineRegisters_Uint16(t *testing.T) {
	regs := []uint16{42}
	val, err := combineRegisters(regs, []int{0}, "uint16")

	require.NoError(t, err)
	assert.Equal(t, float64(42), val)
}

func TestCombineRegisters_Uint16_MaxValue(t *testing.T) {
	regs := []uint16{65535}
	val, err := combineRegisters(regs, []int{0}, "uint16")

	require.NoError(t, err)
	assert.Equal(t, float64(65535), val)
}

func TestCombineRegisters_Int16_Positive(t *testing.T) {
	regs := []uint16{100}
	val, err := combineRegisters(regs, []int{0}, "int16")

	require.NoError(t, err)
	assert.Equal(t, float64(100), val)
}

func TestCombineRegisters_Int16_Negative(t *testing.T) {
	// -1 em int16 = 0xFFFF em uint16
	regs := []uint16{0xFFFF}
	val, err := combineRegisters(regs, []int{0}, "int16")

	require.NoError(t, err)
	assert.Equal(t, float64(-1), val)
}

func TestCombineRegisters_Int16_MinValue(t *testing.T) {
	// -32768 em int16 = 0x8000
	regs := []uint16{0x8000}
	val, err := combineRegisters(regs, []int{0}, "int16")

	require.NoError(t, err)
	assert.Equal(t, float64(-32768), val)
}

func TestCombineRegisters_Float32(t *testing.T) {
	// IEEE 754: 3.14 ≈ 0x4048F5C3
	bits := math.Float32bits(3.14)
	hi := uint16(bits >> 16)
	lo := uint16(bits & 0xFFFF)
	regs := []uint16{hi, lo}

	val, err := combineRegisters(regs, []int{0, 1}, "float32")

	require.NoError(t, err)
	assert.InDelta(t, 3.14, val, 0.001)
}

func TestCombineRegisters_Float32_Zero(t *testing.T) {
	regs := []uint16{0, 0}
	val, err := combineRegisters(regs, []int{0, 1}, "float32")

	require.NoError(t, err)
	assert.Equal(t, float64(0), val)
}

func TestCombineRegisters_Float32_Negative(t *testing.T) {
	bits := math.Float32bits(-273.15)
	hi := uint16(bits >> 16)
	lo := uint16(bits & 0xFFFF)
	regs := []uint16{hi, lo}

	val, err := combineRegisters(regs, []int{0, 1}, "float32")

	require.NoError(t, err)
	assert.InDelta(t, -273.15, val, 0.01)
}

func TestCombineRegisters_Float32Swapped(t *testing.T) {
	bits := math.Float32bits(3.14)
	hi := uint16(bits >> 16)
	lo := uint16(bits & 0xFFFF)
	// Swapped: lo vem primeiro, hi vem depois
	regs := []uint16{lo, hi}

	val, err := combineRegisters(regs, []int{0, 1}, "float32_swapped")

	require.NoError(t, err)
	assert.InDelta(t, 3.14, val, 0.001)
}

func TestCombineRegisters_Int32(t *testing.T) {
	// 100000 = 0x000186A0
	regs := []uint16{0x0001, 0x86A0}
	val, err := combineRegisters(regs, []int{0, 1}, "int32")

	require.NoError(t, err)
	assert.Equal(t, float64(100000), val)
}

func TestCombineRegisters_Int32_Negative(t *testing.T) {
	// -1 em int32 = 0xFFFFFFFF
	regs := []uint16{0xFFFF, 0xFFFF}
	val, err := combineRegisters(regs, []int{0, 1}, "int32")

	require.NoError(t, err)
	assert.Equal(t, float64(-1), val)
}

func TestCombineRegisters_Uint32(t *testing.T) {
	// 70000 = 0x00011170
	regs := []uint16{0x0001, 0x1170}
	val, err := combineRegisters(regs, []int{0, 1}, "uint32")

	require.NoError(t, err)
	assert.Equal(t, float64(70000), val)
}

func TestCombineRegisters_Uint32_MaxValue(t *testing.T) {
	regs := []uint16{0xFFFF, 0xFFFF}
	val, err := combineRegisters(regs, []int{0, 1}, "uint32")

	require.NoError(t, err)
	assert.Equal(t, float64(4294967295), val)
}

func TestCombineRegisters_Auto_1Word(t *testing.T) {
	regs := []uint16{255}
	val, err := combineRegisters(regs, []int{0}, "auto")

	require.NoError(t, err)
	assert.Equal(t, float64(255), val)
}

func TestCombineRegisters_Auto_2Words(t *testing.T) {
	bits := math.Float32bits(99.99)
	hi := uint16(bits >> 16)
	lo := uint16(bits & 0xFFFF)
	regs := []uint16{hi, lo}

	val, err := combineRegisters(regs, []int{0, 1}, "auto")

	require.NoError(t, err)
	assert.InDelta(t, 99.99, val, 0.01)
}

func TestCombineRegisters_Auto_3Words_Error(t *testing.T) {
	regs := []uint16{1, 2, 3}
	_, err := combineRegisters(regs, []int{0, 1, 2}, "auto")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "auto")
}

func TestCombineRegisters_UnknownDataType_Error(t *testing.T) {
	regs := []uint16{42}
	_, err := combineRegisters(regs, []int{0}, "float128")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "float128")
}

func TestCombineRegisters_NonZeroPositionOffset(t *testing.T) {
	// Simula posição dentro de um batch maior de registros
	regs := []uint16{0, 0, 0, 42, 0}
	val, err := combineRegisters(regs, []int{3}, "uint16")

	require.NoError(t, err)
	assert.Equal(t, float64(42), val)
}

// ─── applyOperation ─────────────────────────────────────────────────────────

func ptr(s string) *string { return &s }

func TestApplyOperation_Multiply(t *testing.T) {
	result := applyOperation(10.0, ptr("*"), ptr("3"))
	assert.Equal(t, "30", result)
}

func TestApplyOperation_Divide(t *testing.T) {
	result := applyOperation(100.0, ptr("/"), ptr("4"))
	assert.Equal(t, "25", result)
}

func TestApplyOperation_DivideByZero(t *testing.T) {
	// Divisão por zero deve manter o valor original
	result := applyOperation(100.0, ptr("/"), ptr("0"))
	assert.Equal(t, "100", result)
}

func TestApplyOperation_Add(t *testing.T) {
	result := applyOperation(10.5, ptr("+"), ptr("2.5"))
	assert.Equal(t, "13", result)
}

func TestApplyOperation_Subtract(t *testing.T) {
	result := applyOperation(50.0, ptr("-"), ptr("15"))
	assert.Equal(t, "35", result)
}

func TestApplyOperation_NilOp(t *testing.T) {
	result := applyOperation(42.0, nil, nil)
	assert.Equal(t, "42", result)
}

func TestApplyOperation_NilOpValue(t *testing.T) {
	result := applyOperation(42.0, ptr("*"), nil)
	assert.Equal(t, "42", result)
}

func TestApplyOperation_FloatPrecision(t *testing.T) {
	result := applyOperation(1.0, ptr("/"), ptr("3"))
	// Verifica que o resultado é representado como float
	val := 0.0
	_, err := json.Marshal(result)
	require.NoError(t, err)
	_ = json.Unmarshal([]byte(result), &val)
	assert.InDelta(t, 0.333, val, 0.01)
}

func TestApplyOperation_MultiplyByDecimal(t *testing.T) {
	result := applyOperation(100.0, ptr("*"), ptr("0.1"))
	val := 0.0
	_ = json.Unmarshal([]byte(result), &val)
	assert.InDelta(t, 10.0, val, 0.001)
}

// ─── setAndPublishHealth (state transitions) ────────────────────────────────

func TestSetAndPublishHealth_FirstReport_Publishes(t *testing.T) {
	mock := &MockPublisher{}
	driver := &Driver{
		publisherHealth: mock,
	}

	driver.setAndPublishHealth("dev-001", "online")

	calls := mock.GetHealthCalls()
	require.Len(t, calls, 1)
	assert.Equal(t, "dev-001", calls[0].DeviceID)

	var hp models.HealthPayload
	err := json.Unmarshal([]byte(calls[0].Payload), &hp)
	require.NoError(t, err)
	assert.Equal(t, "online", hp.Status)
	assert.NotEmpty(t, hp.Timestamp)
}

func TestSetAndPublishHealth_SameStatus_DoesNotRepublish(t *testing.T) {
	mock := &MockPublisher{}
	driver := &Driver{
		publisherHealth: mock,
	}

	driver.setAndPublishHealth("dev-001", "online")
	driver.setAndPublishHealth("dev-001", "online")
	driver.setAndPublishHealth("dev-001", "online")

	calls := mock.GetHealthCalls()
	assert.Len(t, calls, 1, "should publish only once when status doesn't change")
}

func TestSetAndPublishHealth_StatusTransition_Publishes(t *testing.T) {
	mock := &MockPublisher{}
	driver := &Driver{
		publisherHealth: mock,
	}

	driver.setAndPublishHealth("dev-001", "online")
	driver.setAndPublishHealth("dev-001", "offline")
	driver.setAndPublishHealth("dev-001", "online")

	calls := mock.GetHealthCalls()
	require.Len(t, calls, 3, "should publish on every state transition")

	var hp1, hp2, hp3 models.HealthPayload
	_ = json.Unmarshal([]byte(calls[0].Payload), &hp1)
	_ = json.Unmarshal([]byte(calls[1].Payload), &hp2)
	_ = json.Unmarshal([]byte(calls[2].Payload), &hp3)

	assert.Equal(t, "online", hp1.Status)
	assert.Equal(t, "offline", hp2.Status)
	assert.Equal(t, "online", hp3.Status)
}

func TestSetAndPublishHealth_MultipleDevices_Independent(t *testing.T) {
	mock := &MockPublisher{}
	driver := &Driver{
		publisherHealth: mock,
	}

	driver.setAndPublishHealth("dev-A", "online")
	driver.setAndPublishHealth("dev-B", "offline")
	driver.setAndPublishHealth("dev-A", "online") // duplicate, skip
	driver.setAndPublishHealth("dev-B", "online") // transition

	calls := mock.GetHealthCalls()
	assert.Len(t, calls, 3, "dev-A second call should be skipped")
}

func TestSetAndPublishHealth_TimestampFormat(t *testing.T) {
	mock := &MockPublisher{}
	driver := &Driver{
		publisherHealth: mock,
	}

	driver.setAndPublishHealth("dev-001", "online")

	calls := mock.GetHealthCalls()
	require.Len(t, calls, 1)

	var hp models.HealthPayload
	err := json.Unmarshal([]byte(calls[0].Payload), &hp)
	require.NoError(t, err)

	_, parseErr := time.Parse(time.RFC3339, hp.Timestamp)
	assert.NoError(t, parseErr, "timestamp should be valid RFC3339")
}

func TestSetAndPublishHealth_ConcurrentAccess(t *testing.T) {
	mock := &MockPublisher{}
	driver := &Driver{
		publisherHealth: mock,
	}

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			status := "online"
			if idx%2 == 0 {
				status = "offline"
			}
			driver.setAndPublishHealth("concurrent-dev", status)
		}(i)
	}
	wg.Wait()

	// Should not panic — validates thread safety via sync.Map
	calls := mock.GetHealthCalls()
	assert.NotEmpty(t, calls)
}
