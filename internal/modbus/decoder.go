package modbus

import (
	"encoding/binary"
	"math"
)

// Sentinel values used by SMA to represent NaN / not measured / open circuit.
const (
	SentinelS32 int32  = -0x80000000
	SentinelU32 uint32 = 0xFFFFFFFF
	SentinelU64 uint64 = 0xFFFFFFFFFFFFFFFF
)

// DecodeS32 converts two 16-bit Big-Endian Modbus registers into a 32-bit signed integer.
// Returns (value, true) if valid, or (0, false) if invalid or equal to SMA sentinel.
func DecodeS32(regs []uint16) (int32, bool) {
	if len(regs) < 2 {
		return 0, false
	}
	b := []byte{
		byte(regs[0] >> 8), byte(regs[0]),
		byte(regs[1] >> 8), byte(regs[1]),
	}
	val := int32(binary.BigEndian.Uint32(b))
	if val == SentinelS32 {
		return 0, false
	}
	return val, true
}

// DecodeU32 converts two 16-bit Big-Endian Modbus registers into a 32-bit unsigned integer.
// Returns (value, true) if valid, or (0, false) if invalid or equal to SMA sentinel.
func DecodeU32(regs []uint16) (uint32, bool) {
	if len(regs) < 2 {
		return 0, false
	}
	b := []byte{
		byte(regs[0] >> 8), byte(regs[0]),
		byte(regs[1] >> 8), byte(regs[1]),
	}
	val := binary.BigEndian.Uint32(b)
	if val == SentinelU32 {
		return 0, false
	}
	return val, true
}

// DecodeU64 converts four 16-bit Big-Endian Modbus registers into a 64-bit unsigned integer.
// Returns (value, true) if valid, or (0, false) if invalid or equal to SMA sentinel.
func DecodeU64(regs []uint16) (uint64, bool) {
	if len(regs) < 4 {
		return 0, false
	}
	b := make([]byte, 8)
	for i := 0; i < 4; i++ {
		b[i*2] = byte(regs[i] >> 8)
		b[i*2+1] = byte(regs[i])
	}
	val := binary.BigEndian.Uint64(b)
	if val == SentinelU64 {
		return 0, false
	}
	return val, true
}

// DecodeOperatingStatus translates an SMA status code into a human-readable string.
func DecodeOperatingStatus(code uint32) string {
	switch code {
	case 307:
		return "OK"
	case 455:
		return "Warning"
	case 35:
		return "Fault"
	case 303:
		return "Off"
	default:
		return "Unknown"
	}
}

// RoundToDecimals rounds float value to specified decimal places.
func RoundToDecimals(val float64, decimals int) float64 {
	pow := math.Pow(10, float64(decimals))
	return math.Round(val*pow) / pow
}
