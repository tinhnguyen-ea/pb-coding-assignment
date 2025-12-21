package repositories

import (
	"context"
	"time"

	"encore.app/billing/domain/entities"
)

type DBRepository interface {
	// GetBillingByExternalID gets a billing by ID
	GetBillingByExternalID(ctx context.Context, externalBillingID string) (*entities.Billing, error)

	// CreateBilling creates a new billing and returns the external billing ID
	CreateBilling(ctx context.Context, userID string, externalBillingID string, description string, currency string, currencyPrecision int64, plannedClosedAt *time.Time) (int64, error)

	// AddLineItem adds a line item to a billing
	AddLineItem(ctx context.Context, billingID int64, description string, amountMinor int64) error

	// CloseBilling closes a billing and sets the actual closed at time
	CloseBilling(ctx context.Context, billingID int64, actualClosedAt time.Time) error

	// CreateBillingSummary creates a billing summary
	CreateBillingSummary(ctx context.Context, externalBillingID string, billingSummary []byte) error

	// GetBillingSummary gets a billing summary
	GetBillingSummary(ctx context.Context, externalBillingID string) (*entities.BillingSummary, error)
}
