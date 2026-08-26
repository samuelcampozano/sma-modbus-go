package modbus

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Client is a thread-safe Modbus TCP client.
type Client struct {
	addr        string
	timeout     time.Duration
	mu          sync.Mutex
	conn        net.Conn
	transID     uint16
}

// NewClient creates a new Modbus TCP client for the specified address.
func NewClient(addr string, timeout time.Duration) *Client {
	return &Client{
		addr:    addr,
		timeout: timeout,
	}
}

// Connect establishes a connection to the Modbus server if not already connected.
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.ensureConnected()
}

func (c *Client) ensureConnected() error {
	if c.conn != nil {
		return nil
	}
	conn, err := net.DialTimeout("tcp", c.addr, c.timeout)
	if err != nil {
		return fmt.Errorf("modbus dial error (%s): %w", c.addr, err)
	}
	c.conn = conn
	return nil
}

// Close terminates the TCP connection.
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		err := c.conn.Close()
		c.conn = nil
		return err
	}
	return nil
}

func (c *Client) reconnect() error {
	if c.conn != nil {
		_ = c.conn.Close()
		c.conn = nil
	}
	return c.ensureConnected()
}

// ReadRegisters reads registers from the target device, trying Holding Registers (0x03) first
// and falling back to Input Registers (0x04) if an exception occurs.
func (c *Client) ReadRegisters(unitID uint8, reg uint16, count uint16) ([]uint16, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.transID++
	currentTransID := c.transID

	// Try Function 0x03 (Read Holding Registers)
	regs, err := c.executeRead(currentTransID, unitID, 0x03, reg, count)
	if err == nil {
		return regs, nil
	}

	// Try Function 0x04 (Read Input Registers) as fallback
	c.transID++
	regsInput, errInput := c.executeRead(c.transID, unitID, 0x04, reg, count)
	if errInput == nil {
		return regsInput, nil
	}

	return nil, err
}

func (c *Client) executeRead(transID uint16, unitID uint8, funcCode uint8, reg uint16, count uint16) ([]uint16, error) {
	if err := c.ensureConnected(); err != nil {
		return nil, err
	}

	req := make([]byte, 12)
	binary.BigEndian.PutUint16(req[0:2], transID)
	binary.BigEndian.PutUint16(req[2:4], 0) // Protocol ID (Modbus = 0)
	binary.BigEndian.PutUint16(req[4:6], 6) // Length (following bytes)
	req[6] = unitID
	req[7] = funcCode
	binary.BigEndian.PutUint16(req[8:10], reg)
	binary.BigEndian.PutUint16(req[10:12], count)

	_ = c.conn.SetDeadline(time.Now().Add(c.timeout))
	if _, err := c.conn.Write(req); err != nil {
		_ = c.reconnect()
		return nil, fmt.Errorf("modbus write error: %w", err)
	}

	// Read MBAP header (7 bytes)
	header := make([]byte, 7)
	if _, err := io.ReadFull(c.conn, header); err != nil {
		_ = c.reconnect()
		return nil, fmt.Errorf("modbus header read error: %w", err)
	}

	respLen := binary.BigEndian.Uint16(header[4:6])
	if respLen < 2 {
		return nil, fmt.Errorf("invalid response length: %d", respLen)
	}

	// Read PDU (length - 1 bytes, since UnitID is byte 7 of header)
	pdu := make([]byte, respLen-1)
	if _, err := io.ReadFull(c.conn, pdu); err != nil {
		_ = c.reconnect()
		return nil, fmt.Errorf("modbus pdu read error: %w", err)
	}

	// Check for Modbus Exception
	respFunc := pdu[0]
	if respFunc&0x80 != 0 {
		return nil, fmt.Errorf("modbus exception: function=0x%02X, code=%d", respFunc, pdu[1])
	}

	byteCount := pdu[1]
	regCount := int(byteCount / 2)
	results := make([]uint16, regCount)
	for i := 0; i < regCount; i++ {
		results[i] = binary.BigEndian.Uint16(pdu[2+i*2 : 4+i*2])
	}
	return results, nil
}
