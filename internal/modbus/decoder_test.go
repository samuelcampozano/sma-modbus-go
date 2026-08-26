package modbus

import (
	"testing"
)

func TestDecodeS32(t *testing.T) {
	tests := []struct {
		name     string
		regs     []uint16
		expected int32
		valid    bool
	}{
		{
			name:     "Positive Power 25000 W",
			regs:     []uint16{0x0000, 0x61A8}, // 25,000
			expected: 25000,
			valid:    true,
		},
		{
			name:     "Negative Value -100",
			regs:     []uint16{0xFFFF, 0xFF9C},
			expected: -100,
			valid:    true,
		},
		{
			name:     "SMA Sentinel NaN (-0x80000000)",
			regs:     []uint16{0x8000, 0x0000},
			expected: 0,
			valid:    false,
		},
		{
			name:     "Short register array",
			regs:     []uint16{0x1234},
			expected: 0,
			valid:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := DecodeS32(tc.regs)
			if ok != tc.valid {
				t.Fatalf("expected valid=%v, got %v", tc.valid, ok)
			}
			if ok && got != tc.expected {
				t.Fatalf("expected %d, got %d", tc.expected, got)
			}
		})
	}
}

func TestDecodeU32(t *testing.T) {
	tests := []struct {
		name     string
		regs     []uint16
		expected uint32
		valid    bool
	}{
		{
			name:     "Frequency 60.00 Hz (6000)",
			regs:     []uint16{0x0000, 0x1770},
			expected: 6000,
			valid:    true,
		},
		{
			name:     "SMA Sentinel 0xFFFFFFFF",
			regs:     []uint16{0xFFFF, 0xFFFF},
			expected: 0,
			valid:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := DecodeU32(tc.regs)
			if ok != tc.valid {
				t.Fatalf("expected valid=%v, got %v", tc.valid, ok)
			}
			if ok && got != tc.expected {
				t.Fatalf("expected %d, got %d", tc.expected, got)
			}
		})
	}
}

func TestDecodeU64(t *testing.T) {
	tests := []struct {
		name     string
		regs     []uint16
		expected uint64
		valid    bool
	}{
		{
			name:     "Energy 678830752 Wh",
			regs:     []uint16{0, 0, 10358, 8864},
			expected: 678830752,
			valid:    true,
		},
		{
			name:     "SMA Sentinel 0xFFFFFFFFFFFFFFFF",
			regs:     []uint16{0xFFFF, 0xFFFF, 0xFFFF, 0xFFFF},
			expected: 0,
			valid:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := DecodeU64(tc.regs)
			if ok != tc.valid {
				t.Fatalf("expected valid=%v, got %v", tc.valid, ok)
			}
			if ok && got != tc.expected {
				t.Fatalf("expected %d, got %d", tc.expected, got)
			}
		})
	}
}

func TestDecodeOperatingStatus(t *testing.T) {
	if DecodeOperatingStatus(307) != "OK" {
		t.Errorf("expected 307 to be 'OK'")
	}
	if DecodeOperatingStatus(455) != "Warning" {
		t.Errorf("expected 455 to be 'Warning'")
	}
	if DecodeOperatingStatus(35) != "Fault" {
		t.Errorf("expected 35 to be 'Fault'")
	}
}
