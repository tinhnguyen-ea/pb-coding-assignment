package activities

import (
	"context"
	"time"

	"encore.dev/rlog"
	"go.temporal.io/sdk/client"

	"encore.app/billing/domain/repositories"
	"encore.app/billing/usecases/dto"
)

type BillingActivities struct {
	dbRepository   repositories.DBRepository
	temporalClient client.Client
	taskQueue      string
}

func NewBillingActivities(
	dbRepository repositories.DBRepository,
	temporalClient client.Client,
	taskQueue string,
) *BillingActivities {
	return &BillingActivities{
		dbRepository:   dbRepository,
		temporalClient: temporalClient,
		taskQueue:      taskQueue,
	}
}

// StartBillingActivity starts a new billing workflow
func (a *BillingActivities) StartBillingActivity(ctx context.Context, userID string, externalBillingID string, description string, currency string, currencyPrecision int64, plannedClosedAt *time.Time) (int64, error) {
	fn := "billingActivities.StartBillingActivity"
	logger := rlog.With("fn", fn).With("userID", userID).With("externalBillingID", externalBillingID).With("description", description).With("currency", currency).With("currencyPrecision", currencyPrecision).With("plannedClosedAt", plannedClosedAt)

	logger.Info("StartBillingActivity starting")

	// Create billing in database
	billingID, err := a.dbRepository.CreateBilling(ctx, userID, externalBillingID, description, currency, currencyPrecision, plannedClosedAt)
	if err != nil {
		logger.Error("Failed to create billing in database", "error", err)
		return 0, dto.ErrFailedToCreateBillingInDatabase
	}

	logger.Info("Billing started in workflow")
	return billingID, nil
}

// AddLineItemActivity adds a line item to a billing
func (a *BillingActivities) AddLineItemActivity(ctx context.Context, billingID int64, description string, amount int64) error {
	fn := "billingActivities.AddLineItemActivity"
	logger := rlog.With("fn", fn).With("billingID", billingID).With("description", description).With("amount", amount)
	logger.Info("AddLineItemActivity starting")

	// Add line item using repository
	err := a.dbRepository.AddLineItem(ctx, billingID, description, amount)
	if err != nil {
		logger.Error("Failed to add line item to database", "error", err)
		return dto.ErrFailedToAddLineItemToDatabase
	}

	logger.Info("Line item added successfully")
	return nil
}

// CloseBillingActivity closes a billing
func (a *BillingActivities) CloseBillingActivity(ctx context.Context, billingID int64) error {
	fn := "billingActivities.CloseBillingActivity"
	logger := rlog.With("fn", fn).With("billingID", billingID)

	logger.Info("CloseBillingActivity starting")

	// close billing in database
	actualClosedAt := time.Now().UTC()
	err := a.dbRepository.CloseBilling(ctx, billingID, actualClosedAt)
	if err != nil {
		logger.Error("Failed to close billing in database", "error", err)
		return err
	}

	logger.Info("Billing closed successfully")
	return nil
}

// CreateBillingSummaryActivity creates a billing summary
func (a *BillingActivities) CreateBillingSummaryActivity(ctx context.Context, externalBillingID string, billingSummary []byte) error {
	fn := "billingActivities.CreateBillingSummaryActivity"
	logger := rlog.With("fn", fn).With("externalBillingID", externalBillingID)

	logger.Info("CreateBillingSummaryActivity starting")

	// generate billing summary in database
	err := a.dbRepository.CreateBillingSummary(ctx, externalBillingID, billingSummary)
	if err != nil {
		logger.Error("Failed to generate billing summary in database", "error", err)
		return err
	}

	return nil
}

// Package-level activity functions for type-safe workflow references
var (
	// Activity instances are set during initialization
	activityInstance *BillingActivities
)

// SetActivityInstance sets the activity instance for package-level functions
func SetActivityInstance(activities *BillingActivities) {
	activityInstance = activities
}

// StartBillingActivityFunc is a package-level function wrapper for StartBillingActivity
func StartBillingActivityFunc(ctx context.Context, userID string, externalBillingID string, description string, currency string, currencyPrecision int64, plannedClosedAt *time.Time) (int64, error) {
	if activityInstance == nil {
		panic("activity instance not initialized - call SetActivityInstance first")
	}
	return activityInstance.StartBillingActivity(ctx, userID, externalBillingID, description, currency, currencyPrecision, plannedClosedAt)
}

// AddLineItemActivityFunc is a package-level function wrapper for AddLineItemActivity
func AddLineItemActivityFunc(ctx context.Context, billingID int64, description string, amount int64) error {
	if activityInstance == nil {
		panic("activity instance not initialized - call SetActivityInstance first")
	}
	return activityInstance.AddLineItemActivity(ctx, billingID, description, amount)
}

// CloseBillingActivityFunc is a package-level function wrapper for CloseBillingActivity
func CloseBillingActivityFunc(ctx context.Context, billingID int64) error {
	if activityInstance == nil {
		panic("activity instance not initialized - call SetActivityInstance first")
	}
	return activityInstance.CloseBillingActivity(ctx, billingID)
}

// CreateBillingSummaryActivityFunc is a package-level function wrapper for CreateBillingSummaryActivity
func CreateBillingSummaryActivityFunc(ctx context.Context, externalBillingID string, billingSummary []byte) error {
	if activityInstance == nil {
		panic("activity instance not initialized - call SetActivityInstance first")
	}
	return activityInstance.CreateBillingSummaryActivity(ctx, externalBillingID, billingSummary)
}
