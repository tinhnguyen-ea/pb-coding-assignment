package services

import (
	"context"
	"time"

	"encore.app/billing/domain/entities"
)

type FxService = interface {
	GetSupportedCurrencies(ctx context.Context, requestTime time.Time) ([]string, error)
	GetCurrencyMetadata(ctx context.Context, currency string, requestTime time.Time) (*entities.CurrencyMetadata, error)
	GetRates(ctx context.Context, requestTime time.Time) (*map[string]entities.CurrencyRate, error)
}
