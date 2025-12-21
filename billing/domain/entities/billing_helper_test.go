package entities

import (
	"testing"
)

func TestHasAtMostXDecimals(t *testing.T) {
	tests := []struct {
		name     string
		f        float64
		x        int64
		expected bool
	}{
		{
			name:     "USD amount with 2 decimals",
			f:        10.99,
			x:        2,
			expected: true,
		},
		{
			name:     "USD amount with 3 decimals (invalid)",
			f:        10.999,
			x:        2,
			expected: false,
		},
		{
			name:     "JPY amount with 0 decimals",
			f:        100.0,
			x:        0,
			expected: true,
		},
		{
			name:     "JPY amount with 1 decimal (invalid)",
			f:        100.5,
			x:        0,
			expected: false,
		},
		{
			name:     "amount with trailing zeros",
			f:        10.00,
			x:        2,
			expected: true,
		},
		{
			name:     "amount with trailing zeros, precision 1",
			f:        10.00,
			x:        1,
			expected: true, // 10.0 is valid with precision 1
		},
		{
			name:     "very small number",
			f:        0.0000001,
			x:        7,
			expected: true,
		},
		{
			name:     "very small number with less precision",
			f:        0.0000001,
			x:        6,
			expected: false,
		},
		{
			name:     "number with many trailing zeros",
			f:        10.0000000,
			x:        2,
			expected: true,
		},
		{
			name:     "zero value",
			f:        0.0,
			x:        0,
			expected: true,
		},
		{
			name:     "very large number",
			f:        999999999.99,
			x:        2,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasAtMostXDecimals(tt.f, tt.x)
			if result != tt.expected {
				t.Errorf("hasAtMostXDecimals(%f, %d) = %v, expected %v", tt.f, tt.x, result, tt.expected)
			}
		})
	}
}
