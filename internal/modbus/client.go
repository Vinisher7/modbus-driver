package modbus

import (
	"encoding/binary"
	"fmt"
	"time"

	"github.com/goburrow/modbus"
)

type Client struct {
	handler *modbus.TCPClientHandler
	client  modbus.Client
}

func NewClient(host string, port, unitID int, timeout time.Duration) (*Client, error) {
	addr := fmt.Sprintf("%s:%d", host, port)
	h := modbus.NewTCPClientHandler(addr)
	h.Timeout = timeout
	h.SlaveId = byte(unitID)

	if err := h.Connect(); err != nil {
		return nil, fmt.Errorf("modbus connect %s: %w", addr, err)
	}
	return &Client{handler: h, client: modbus.NewClient(h)}, nil
}

func (c *Client) Close() {
	_ = c.handler.Close()
}

// ReadRegisters lê `quantity` holding registers a partir de `address`
// e retorna um slice de uint16
func (c *Client) ReadRegisters(functionCode, address, quantity int) ([]uint16, error) {
	var raw []byte
	var err error

	switch functionCode {
	case 1:
		raw, err = c.client.ReadCoils(uint16(address), uint16(quantity))
	case 2:
		raw, err = c.client.ReadDiscreteInputs(uint16(address), uint16(quantity))
	case 3:
		raw, err = c.client.ReadHoldingRegisters(uint16(address), uint16(quantity))
	case 4:
		raw, err = c.client.ReadInputRegisters(uint16(address), uint16(quantity))
	default:
		return nil, fmt.Errorf("function code %d not supported", functionCode)
	}
	if err != nil {
		return nil, err
	}

	regs := make([]uint16, len(raw)/2)
	for i := range regs {
		regs[i] = binary.BigEndian.Uint16(raw[i*2:])
	}
	return regs, nil
}
