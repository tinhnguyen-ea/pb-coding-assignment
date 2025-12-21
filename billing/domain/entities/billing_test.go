package entities

import (
	"testing"
	"time"
)

func TestBilling_CanAddLineItem(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		billing  *Billing
		expected bool
	}{
		{
			name: "open status can add line item",
			billing: &Billing{
				Status: BillingStatusOpen,
			},
			expected: true,
		},
		{
			name: "closed status cannot add line item",
			billing: &Billing{
				Status:         BillingStatusClosed,
				ActualClosedAt: &now,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.billing.CanAddLineItem()
			if result != tt.expected {
				t.Errorf("CanAddLineItem() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestBilling_CanCloseBilling(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		billing  *Billing
		expected bool
	}{
		{
			name: "open status can close billing",
			billing: &Billing{
				Status: BillingStatusOpen,
			},
			expected: true,
		},
		{
			name: "closed status cannot close billing",
			billing: &Billing{
				Status:         BillingStatusClosed,
				ActualClosedAt: &now,
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.billing.CanCloseBilling()
			if result != tt.expected {
				t.Errorf("CanCloseBilling() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestBilling_CanAddItemWithAmount(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		billing  *Billing
		amount   float64
		expected bool
	}{
		{
			name: "open status with valid amount",
			billing: &Billing{
				Status:            BillingStatusOpen,
				CurrencyPrecision: 2,
			},
			amount:   10.99,
			expected: true,
		},
		{
			name: "open status with invalid amount (too many decimals)",
			billing: &Billing{
				Status:            BillingStatusOpen,
				CurrencyPrecision: 2,
			},
			amount:   10.999,
			expected: false,
		},
		{
			name: "closed status with valid amount",
			billing: &Billing{
				Status:            BillingStatusClosed,
				CurrencyPrecision: 2,
				ActualClosedAt:    &now,
			},
			amount:   10.99,
			expected: false,
		},
		{
			name: "open status with zero amount",
			billing: &Billing{
				Status:            BillingStatusOpen,
				CurrencyPrecision: 2,
			},
			amount:   0.0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.billing.CanAddItemWithAmount(tt.amount)
			if result != tt.expected {
				t.Errorf("CanAddItemWithAmount(%f) = %v, expected %v", tt.amount, result, tt.expected)
			}
		})
	}
}
