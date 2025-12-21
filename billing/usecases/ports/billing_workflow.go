package ports

import (
	"context"
	"time"

	"encore.app/billing/domain/entities"
)

type BillingWorkflow interface {
	// StartBilling starts a billing
	StartBilling(ctx context.Context, userID string, billingID string, description string, currency string, currencyPrecision int64, plannedClosedAt *time.Time) error

	// AddLineItem adds a line item to a billing
	// billingID is the internal auto-incremented billing ID
	AddLineItem(ctx context.Context, externalBillingID string, description string, amountMinor int64) error

	// CloseBilling closes a billing
	CloseBilling(ctx context.Context, externalBillingID string) error

	// GetBillingSummary gets a billing summary
	GetBillingSummary(ctx context.Context, externalBillingID string) (*entities.BillingSummary, error)
}
