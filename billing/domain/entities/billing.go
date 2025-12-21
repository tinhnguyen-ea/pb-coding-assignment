package entities

import (
	"time"
)

type BillingStatus = string

const (
	BillingStatusOpen           BillingStatus = "open"
	BillingStatusPendingClosure BillingStatus = "pending_closure"
	BillingStatusClosed         BillingStatus = "closed"
)

type Billing struct {
	ID                int64         `json:"id"`
	UserID            string        `json:"user_id"`
	Description       string        `json:"description"`
	Currency          string        `json:"currency"`
	CurrencyPrecision int64         `json:"currency_precision"`
	Status            BillingStatus `json:"status"`
	PlannedClosedAt   *time.Time    `json:"planned_closed_at"`
	ActualClosedAt    *time.Time    `json:"actual_closed_at"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

func (b *Billing) CanAddLineItem() bool {
	return b.Status == BillingStatusOpen
}

func (b *Billing) CanCloseBilling() bool {
	return b.Status == BillingStatusOpen
}

func (b *Billing) CanAddItemWithAmount(amount float64) bool {
	return b.CanAddLineItem() && hasAtMostXDecimals(amount, b.CurrencyPrecision)
}

type LineItem struct {
	Description string `json:"description"`
	AmountMinor int64  `json:"amount_minor"`
}

type BillingSummary struct {
	ExternalBillingID string     `json:"external_billing_id"`
	Description       string     `json:"description"`
	Currency          string     `json:"currency"`
	CurrencyPrecision int64      `json:"currency_precision"`
	LineItems         []LineItem `json:"line_items"`
	TotalAmountMinor  int64      `json:"total_amount_minor"`
}
