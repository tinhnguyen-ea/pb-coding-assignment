package billing

import "time"

type CreateBillingRequest struct {
	UserID      string `json:"user_id"`
	Description string `json:"description"` // optional
	Currency    string `json:"currency"`

	// PlannedClosedAt is designed for periodic billing (e.g. weekly, monthly, etc). If not provided, the billing will not be closed automatically.
	PlannedClosedAt *time.Time `json:"planned_closed_at,omitempty"`
}

type CreateBillingResponse struct {
	BillingID string `json:"billing_id"`
}

type AddLineItemRequest struct {
	Description string  `json:"description"`
	Amount      float64 `json:"amount"`
}

type LineItem struct {
	Description string `json:"description"`
	AmountMinor int64  `json:"amountMinor"`
}

type GetBillingSummaryResponse struct {
	ExternalBillingID string     `json:"billing_id"`
	Description       string     `json:"description"`
	Currency          string     `json:"currency"`
	CurrencyPrecision int64      `json:"currency_precision"`
	LineItems         []LineItem `json:"line_items"`
	TotalAmountMinor  int64      `json:"total_amount_minor"`
}
