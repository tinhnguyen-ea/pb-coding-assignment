package entities

type CurrencyMetadata struct {
	Code      string `json:"code"`
	Symbol    string `json:"symbol"`
	Precision int64  `json:"precision"`
}

type CurrencyRate struct {
	Rate      int64 `json:"rate"`
	Precision int64 `json:"precision"`
}
