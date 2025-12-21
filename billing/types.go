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
