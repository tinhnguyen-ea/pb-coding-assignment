package services

import (
	"context"
	"errors"
	"testing"
	"time"

	"encore.app/billing/domain/entities"
	"encore.app/fx"
	"encore.dev/et"
)

func TestNewFxService(t *testing.T) {
	service := NewFxService()
	if service == nil {
		t.Error("NewFxService returned nil")
	}
}

func TestFxService_GetSupportedCurrencies(t *testing.T) {
	ctx := context.Background()
	requestTime := time.Now()

	// Mock the fx.GetSupportedCurrencies endpoint
	et.MockEndpoint(fx.GetSupportedCurrencies, func(ctx context.Context, req *fx.GetSupportedCurrenciesRequest) (*fx.GetSupportedCurrenciesResponse, error) {
		return &fx.GetSupportedCurrenciesResponse{
			Currencies: []string{"USD", "GEL", "EUR"},
		}, nil
	})

	service := NewFxService()
	currencies, err := service.GetSupportedCurrencies(ctx, requestTime)
	if err != nil {
		t.Fatalf("GetSupportedCurrencies failed: %v", err)
	}
	if len(currencies) != 3 {
		t.Errorf("Expected 3 currencies, got %d", len(currencies))
	}
	if currencies[0] != "USD" {
		t.Errorf("Expected first currency USD, got %s", currencies[0])
	}
}

func TestFxService_GetSupportedCurrencies_Error(t *testing.T) {
	ctx := context.Background()
	requestTime := time.Now()

	// Mock the fx.GetSupportedCurrencies endpoint to return error
	et.MockEndpoint(fx.GetSupportedCurrencies, func(ctx context.Context, req *fx.GetSupportedCurrenciesRequest) (*fx.GetSupportedCurrenciesResponse, error) {
		return nil, entities.ErrFxService
	})

	service := NewFxService()
	currencies, err := service.GetSupportedCurrencies(ctx, requestTime)
	if err == nil {
		t.Error("Expected error from GetSupportedCurrencies")
	}
	if !errors.Is(err, entities.ErrFxService) {
		t.Errorf("Expected ErrFxService, got: %v", err)
	}
	if len(currencies) != 0 {
		t.Errorf("Expected empty currencies on error, got %d", len(currencies))
	}
}

func TestFxService_GetCurrencyMetadata(t *testing.T) {
	ctx := context.Background()
	requestTime := time.Now()

	// Mock the fx.GetCurrencyMetadata endpoint
	et.MockEndpoint(fx.GetCurrencyMetadata, func(ctx context.Context, req *fx.CurrencyMetadataRequest) (*fx.CurrencyMetadataResponse, error) {
		return &fx.CurrencyMetadataResponse{
			Metadata: map[string]fx.CurrencyMetadata{
				"USD": {
					Code:      "USD",
					Symbol:    "$",
					Precision: 2,
				},
			},
		}, nil
	})

	service := NewFxService()
	metadata, err := service.GetCurrencyMetadata(ctx, "USD", requestTime)
	if err != nil {
		t.Fatalf("GetCurrencyMetadata failed: %v", err)
	}
	if metadata.Code != "USD" {
		t.Errorf("Expected code USD, got %s", metadata.Code)
	}
	if metadata.Symbol != "$" {
		t.Errorf("Expected symbol $, got %s", metadata.Symbol)
	}
	if metadata.Precision != 2 {
		t.Errorf("Expected precision 2, got %d", metadata.Precision)
	}
}

func TestFxService_GetCurrencyMetadata_Error(t *testing.T) {
	ctx := context.Background()
	requestTime := time.Now()

	// Mock the fx.GetCurrencyMetadata endpoint to return error
	et.MockEndpoint(fx.GetCurrencyMetadata, func(ctx context.Context, req *fx.CurrencyMetadataRequest) (*fx.CurrencyMetadataResponse, error) {
		return nil, entities.ErrFxService
	})

	service := NewFxService()
	metadata, err := service.GetCurrencyMetadata(ctx, "USD", requestTime)
	if err == nil {
		t.Error("Expected error from GetCurrencyMetadata")
	}
	if !errors.Is(err, entities.ErrFxService) {
		t.Errorf("Expected ErrFxService, got: %v", err)
	}
	if metadata.Code != "" {
		t.Errorf("Expected empty metadata on error, got code %s", metadata.Code)
	}
}

func TestFxService_GetRates(t *testing.T) {
	ctx := context.Background()
	requestTime := time.Now()

	// Mock the fx.GetRates endpoint
	et.MockEndpoint(fx.GetRates, func(ctx context.Context, req *fx.CurrencyRatesRequest) (*fx.CurrencyRatesResponse, error) {
		return &fx.CurrencyRatesResponse{
			Rates: map[string]fx.CurrencyRate{
				"USD": {
					Rate:      100,
					Precision: 2,
				},
				"GEL": {
					Rate:      270,
					Precision: 2,
				},
			},
		}, nil
	})

	service := NewFxService()
	rates, err := service.GetRates(ctx, requestTime)
	if err != nil {
		t.Fatalf("GetRates failed: %v", err)
	}
	if rates == nil {
		t.Error("GetRates returned nil rates")
	}
	if len(*rates) != 2 {
		t.Errorf("Expected 2 rates, got %d", len(*rates))
	}
	usdRate, ok := (*rates)["USD"]
	if !ok {
		t.Error("USD rate not found")
	}
	if usdRate.Rate != 100 {
		t.Errorf("Expected USD rate 100, got %d", usdRate.Rate)
	}
}

func TestFxService_GetRates_Error(t *testing.T) {
	ctx := context.Background()
	requestTime := time.Now()

	// Mock the fx.GetRates endpoint to return error
	et.MockEndpoint(fx.GetRates, func(ctx context.Context, req *fx.CurrencyRatesRequest) (*fx.CurrencyRatesResponse, error) {
		return nil, entities.ErrFxService
	})

	service := NewFxService()
	rates, err := service.GetRates(ctx, requestTime)
	if err == nil {
		t.Error("Expected error from GetRates")
	}
	if !errors.Is(err, entities.ErrFxService) {
		t.Errorf("Expected ErrFxService, got: %v", err)
	}
	if rates == nil || len(*rates) != 0 {
		t.Error("Expected empty rates on error")
	}
}
