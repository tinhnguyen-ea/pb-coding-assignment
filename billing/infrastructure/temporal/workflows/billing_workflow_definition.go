package workflows

import (
	"encoding/json"
	"time"

	"encore.dev/rlog"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"encore.app/billing/infrastructure/temporal/activities"
)

const (
	AddLineItemSignal  = "add-line-item"
	CloseBillingSignal = "close-billing"
)

type BillingWorkflowInput struct {
	UserID            string     `json:"user_id"`
	ExternalBillingID string     `json:"billing_id"`
	Description       string     `json:"description"`
	Currency          string     `json:"currency"`
	CurrencyPrecision int64      `json:"currency_precision"`
	PlannedClosedAt   *time.Time `json:"planned_closed_at"`
}

type BillingWorkflowState struct {
	ExternalBillingID string          `json:"external_billing_id"`
	BillingID         int64           `json:"-"`
	Description       string          `json:"description"`
	Currency          string          `json:"currency"`
	CurrencyPrecision int64           `json:"currency_precision"`
	Status            string          `json:"-"`
	LineItems         []LineItemState `json:"line_items"`
	ClosedAt          *time.Time      `json:"-"`
	LastActivity      time.Time       `json:"-"`
	TotalAmountMinor  int64           `json:"total_amount_minor"`
}

type LineItemState struct {
	Description string    `json:"description"`
	AmountMinor int64     `json:"amount_minor"`
	AddedAt     time.Time `json:"added_at"`
}

// BillingWorkflow is the Temporal workflow for managing billing lifecycle
func BillingWorkflow(ctx workflow.Context, input BillingWorkflowInput) error {
	fn := "billingWorkflowDefinition.BillingWorkflow"
	logger := rlog.With("fn", fn).With("externalBillingID", input.ExternalBillingID).With("userID", input.UserID).With("description", input.Description).With("currency", input.Currency).With("currencyPrecision", input.CurrencyPrecision).With("plannedClosedAt", input.PlannedClosedAt)

	logger.Info("BillingWorkflow starting")

	// Initialize workflow state
	state := BillingWorkflowState{
		ExternalBillingID: input.ExternalBillingID,
		Currency:          input.Currency,
		CurrencyPrecision: input.CurrencyPrecision,
		Description:       input.Description,
		Status:            "open",
		LineItems:         []LineItemState{},
		LastActivity:      workflow.Now(ctx),
		TotalAmountMinor:  0,
	}

	// Activity options
	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    10,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// set query handler for current state
	err := workflow.SetQueryHandler(ctx, "currentState", func() (BillingWorkflowState, error) {
		return state, nil
	})
	if err != nil {
		logger.Error("Failed to set query handler", "error", err)
		return err
	}

	// start billing activity
	var billingID int64
	err = workflow.ExecuteActivity(ctx, activities.StartBillingActivityFunc, input.UserID, input.ExternalBillingID, input.Description, input.Currency, input.CurrencyPrecision, input.PlannedClosedAt).Get(ctx, &billingID)
	if err != nil {
		logger.Error("Failed to start billing", "error", err)
		return err
	}

	// update internal billingID in state
	state.BillingID = billingID

	// Helper function to close billing and generate summary
	closeBillingAndGenerateSummary := func() {
		logger.Info("Closing billing")

		// Execute activity to close billing
		err := workflow.ExecuteActivity(ctx, activities.CloseBillingActivityFunc, state.BillingID).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to close billing", "error", err)
			return
		}

		// Generate billing summary
		billingSummary, err := json.Marshal(state)
		if err != nil {
			logger.Error("Failed to generate billing summary", "error", err)
			return
		}

		err = workflow.ExecuteActivity(ctx, activities.CreateBillingSummaryActivityFunc, input.ExternalBillingID, billingSummary).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to generate billing summary", "error", err)
			return
		}

		now := workflow.Now(ctx)
		state.ClosedAt = &now
		state.Status = "closed"
		state.LastActivity = now

		logger.Info("Billing closed and summary generated")
	}

	// Wait for line items to be added or billing to be closed
	selector := workflow.NewSelector(ctx)

	// Channel for line item additions
	lineItemChan := workflow.GetSignalChannel(ctx, AddLineItemSignal)

	// Channel for closing billing (manual close)
	closeChan := workflow.GetSignalChannel(ctx, CloseBillingSignal)

	// Timer for auto-close at plannedClosedAt (if set)
	var autoCloseTimer workflow.Future
	if input.PlannedClosedAt != nil {
		now := workflow.Now(ctx)
		duration := input.PlannedClosedAt.Sub(now)
		autoCloseTimer = workflow.NewTimer(ctx, duration)
	}

	selector.AddReceive(lineItemChan, func(c workflow.ReceiveChannel, more bool) {
		var lineItem LineItemState
		c.Receive(ctx, &lineItem)
		logger.Info("Received add line item signal", "description", lineItem.Description, "amountMinor", lineItem.AmountMinor)

		// Execute activity to add line item
		err := workflow.ExecuteActivity(ctx, activities.AddLineItemActivityFunc, state.BillingID, lineItem.Description, lineItem.AmountMinor).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to add line item", "error", err)
			return
		}

		// Update state
		state.LineItems = append(state.LineItems, lineItem)
		state.TotalAmountMinor = state.TotalAmountMinor + lineItem.AmountMinor
		state.LastActivity = workflow.Now(ctx)
	})

	selector.AddReceive(closeChan, func(c workflow.ReceiveChannel, more bool) {
		var closeSignal struct{}
		c.Receive(ctx, &closeSignal)
		logger.Info("Received manual close billing signal")

		closeBillingAndGenerateSummary()
	})

	// Add auto-close timer to selector if it exists
	if autoCloseTimer != nil {
		selector.AddFuture(autoCloseTimer, func(f workflow.Future) {
			// Check if billing is already closed (manual close may have happened)
			if state.Status == "closed" {
				logger.Info("Auto-close timer fired but billing already closed")
				return
			}

			logger.Info("Auto-close timer fired", "plannedClosedAt", input.PlannedClosedAt)

			closeBillingAndGenerateSummary()
		})
	}

	// Wait for signals
	for state.Status != "closed" {
		selector.Select(ctx)
	}

	logger.Info("BillingWorkflow completed", "billingID", input.ExternalBillingID)
	return nil
}
