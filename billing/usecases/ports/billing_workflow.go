package ports

import (
	"context"
	"time"
)

type BillingWorkflow interface {
	// StartBilling starts a billing
	StartBilling(ctx context.Context, userID string, billingID string, description string, currency string, currencyPrecision int64, plannedClosedAt *time.Time) error

	// AddLineItem adds a line item to a billing
	// billingID is the internal auto-incremented billing ID
	AddLineItem(ctx context.Context, externalBillingID string, description string, amountMinor int64) error

	// CloseBilling closes a billing
	CloseBilling(ctx context.Context, externalBillingID string) error
}
