package modbus

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ─── ReadRegisters function code routing ─────────────────────────────────────

func TestReadRegisters_UnsupportedFunctionCode(t *testing.T) {
	// We can't easily test with a real Modbus server, but we CAN test
	// the error path for an unsupported function code without any connection.
	c := &Client{} // nil handler/client — we won't reach the network call

	_, err := c.ReadRegisters(99, 0, 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "function code 99 not supported")
}

func TestReadRegisters_FunctionCode0_Unsupported(t *testing.T) {
	c := &Client{}

	_, err := c.ReadRegisters(0, 0, 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "function code 0 not supported")
}

func TestReadRegisters_FunctionCode5_Unsupported(t *testing.T) {
	c := &Client{}

	_, err := c.ReadRegisters(5, 0, 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "function code 5 not supported")
}

func TestReadRegisters_NegativeFunctionCode(t *testing.T) {
	c := &Client{}

	_, err := c.ReadRegisters(-1, 0, 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
}
