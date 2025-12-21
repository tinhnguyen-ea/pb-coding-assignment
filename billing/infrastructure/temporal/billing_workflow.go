package temporal

import (
	"context"
	"fmt"
	"time"

	"encore.dev/rlog"
	"go.temporal.io/sdk/client"

	"encore.app/billing/infrastructure/temporal/workflows"
	"encore.app/billing/usecases/dto"
	"encore.app/billing/usecases/ports"
)

const (
	BillingWorkflowName = "billing-workflow"
	WorkflowTaskQueue   = "billing-workflow-task-queue"
	WorkflowIDPrefix    = "billing-workflow-"
)

type TemporalBillingWorkflow struct {
	client    client.Client
	taskQueue string
}

func NewTemporalBillingWorkflow(client client.Client, taskQueue string) ports.BillingWorkflow {
	return &TemporalBillingWorkflow{
		client:    client,
		taskQueue: taskQueue,
	}
}

// StartBilling starts a billing workflow
func (s *TemporalBillingWorkflow) StartBilling(ctx context.Context, userID string, externalBillingID string, description string, currency string, currencyPrecision int64, plannedClosedAt *time.Time) error {
	logger := rlog.With("fn", "TemporalBillingWorkflow.StartBill").With("userID", userID).With("externalBillingID", externalBillingID).With("description", description).With("currency", currency).With("currencyPrecision", currencyPrecision).With("plannedClosedAt", plannedClosedAt)

	workflowID := fmt.Sprintf("%s%s", WorkflowIDPrefix, externalBillingID)
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: s.taskQueue,
	}

	input := workflows.BillingWorkflowInput{
		UserID:            userID,
		ExternalBillingID: externalBillingID,
		Description:       description,
		Currency:          currency,
		CurrencyPrecision: currencyPrecision,
		PlannedClosedAt:   plannedClosedAt,
	}

	logger.Info("Starting billing workflow", "workflowID", workflowID)

	_, err := s.client.ExecuteWorkflow(ctx, workflowOptions, workflows.BillingWorkflow, input)
	if err != nil {
		logger.Error("Failed to start billing workflow", "error", err)
		return dto.ErrFailedToStartBillingWorkflow
	}

	logger.Info("Billing workflow started", "workflowID", workflowID)
	return nil
}

// AddLineItem sends a signal to add a line item to the billing workflow
func (s *TemporalBillingWorkflow) AddLineItem(ctx context.Context, externalBillingID string, description string, amountMinor int64) error {
	logger := rlog.With("fn", "TemporalBillingWorkflow.AddLineItem").With("externalBillingID", externalBillingID).With("description", description).With("amountMinor", amountMinor)

	workflowID := fmt.Sprintf("%s%s", WorkflowIDPrefix, externalBillingID)

	lineItem := workflows.LineItemState{
		Description: description,
		AmountMinor: amountMinor,
		AddedAt:     time.Now().UTC(),
	}

	err := s.client.SignalWorkflow(ctx, workflowID, "", workflows.AddLineItemSignal, lineItem)
	if err != nil {
		logger.Error("Failed to signal add-line-item", "error", err)
		return fmt.Errorf("failed to signal add-line-item: %w", err)
	}

	logger.Info("Add-line-item signal sent", "workflowID", workflowID)
	return nil
}

// CloseBilling sends a signal to close the billing workflow
func (s *TemporalBillingWorkflow) CloseBilling(ctx context.Context, externalBillingID string) error {
	logger := rlog.With("fn", "TemporalBillingWorkflow.CloseBilling").With("externalBillingID", externalBillingID)

	workflowID := fmt.Sprintf("%s%s", WorkflowIDPrefix, externalBillingID)

	err := s.client.SignalWorkflow(ctx, workflowID, "", workflows.CloseBillingSignal, struct{}{})
	if err != nil {
		logger.Error("Failed to signal close-billing", "error", err)
		return fmt.Errorf("failed to signal close-billing: %w", err)
	}

	logger.Info("Close-billing signal sent", "workflowID", workflowID)
	return nil
}
