package fx

import (
	"context"
	"time"
)

var rates = map[string]CurrencyRate{
	"USD": { // rate of 1 USD to 100 minor units
		Rate:      100,
		Precision: 2,
	},
	"GEL": {
		// rate of 1 USD to 270 minor units (1 USD = 2.70 GEL)
		Rate:      270,
		Precision: 2,
	},
}

var currencyMetadata = map[string]CurrencyMetadata{
	"USD": {
		Code:      "USD",
		Symbol:    "$",
		Precision: 2,
	},
	"GEL": {
		Code:      "GEL",
		Symbol:    "₾",
		Precision: 2,
	},
}

type CurrencyRate struct {
	Rate      int64 `json:"rate"`
	Precision int64 `json:"precision"`
}

type CurrencyMetadata struct {
	Code      string `json:"code"`
	Symbol    string `json:"symbol"`
	Precision int64  `json:"precision"`
}

type GetSupportedCurrenciesRequest struct {
	RequestTime time.Time `json:"request_time"`
}

type GetSupportedCurrenciesResponse struct {
	Currencies []string `json:"currencies"`
}

type CurrencyRatesRequest struct {
	RequestTime time.Time `json:"request_time"`
}

type CurrencyRatesResponse struct {
	Rates map[string]CurrencyRate `json:"rates"`
}

type CurrencyMetadataRequest struct {
	RequestTime time.Time `json:"request_time"`
}

type CurrencyMetadataResponse struct {
	Metadata map[string]CurrencyMetadata `json:"metadata"`
}

//encore:api private method=GET path=/fx/supported-currencies
func GetSupportedCurrencies(ctx context.Context, req *GetSupportedCurrenciesRequest) (*GetSupportedCurrenciesResponse, error) {
	return &GetSupportedCurrenciesResponse{
		Currencies: []string{"USD", "GEL"},
	}, nil
}

//encore:api private method=GET path=/fx/metadata
func GetCurrencyMetadata(ctx context.Context, req *CurrencyMetadataRequest) (*CurrencyMetadataResponse, error) {
	return &CurrencyMetadataResponse{
		Metadata: currencyMetadata,
	}, nil
}

//encore:api private method=GET path=/fx/rates
func GetRates(ctx context.Context, req *CurrencyRatesRequest) (*CurrencyRatesResponse, error) {
	return &CurrencyRatesResponse{
		Rates: rates,
	}, nil
}
