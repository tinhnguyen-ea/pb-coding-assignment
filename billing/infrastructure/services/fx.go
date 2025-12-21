package services

import (
	"context"
	"time"

	"encore.app/billing/domain/entities"
	"encore.app/billing/domain/services"
	"encore.app/fx"
)

type fxService struct{}

func NewFxService() services.FxService {
	return &fxService{}
}

func (s *fxService) GetSupportedCurrencies(ctx context.Context, requestTime time.Time) ([]string, error) {
	supportedCurrencies, err := fx.GetSupportedCurrencies(ctx, &fx.GetSupportedCurrenciesRequest{
		RequestTime: requestTime,
	})
	if err != nil {
		return []string{}, entities.ErrFxService
	}
	return supportedCurrencies.Currencies, nil
}

func (s *fxService) GetCurrencyMetadata(ctx context.Context, currency string, requestTime time.Time) (*entities.CurrencyMetadata, error) {
	currencyMetadata, err := fx.GetCurrencyMetadata(ctx, &fx.CurrencyMetadataRequest{
		RequestTime: requestTime,
	})
	if err != nil {
		return &entities.CurrencyMetadata{}, entities.ErrFxService
	}
	return &entities.CurrencyMetadata{
		Code:      currencyMetadata.Metadata[currency].Code,
		Symbol:    currencyMetadata.Metadata[currency].Symbol,
		Precision: currencyMetadata.Metadata[currency].Precision,
	}, nil
}

func (s *fxService) GetRates(ctx context.Context, requestTime time.Time) (*map[string]entities.CurrencyRate, error) {
	rates, err := fx.GetRates(ctx, &fx.CurrencyRatesRequest{
		RequestTime: requestTime,
	})
	if err != nil {
		return &map[string]entities.CurrencyRate{}, entities.ErrFxService
	}

	currencyRates := make(map[string]entities.CurrencyRate)
	for currency, rate := range rates.Rates {
		currencyRates[currency] = entities.CurrencyRate{
			Rate:      rate.Rate,
			Precision: rate.Precision,
		}
	}
	return &currencyRates, nil
}
