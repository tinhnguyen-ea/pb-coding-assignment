package repositories

import (
	"context"
	"time"
)

type DBRepository interface {
	// CreateBilling creates a new billing and returns the external billing ID
	CreateBilling(ctx context.Context, userID string, externalBillingID string, description string, currency string, currencyPrecision int64, plannedClosedAt *time.Time) (string, error)
}
