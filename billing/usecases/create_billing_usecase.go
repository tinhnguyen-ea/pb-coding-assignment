package usecases

import (
	"context"
	"slices"
	"time"

	"encore.dev/rlog"
	"github.com/google/uuid"

	"encore.app/billing/domain/services"
	"encore.app/billing/usecases/dto"
	"encore.app/billing/usecases/ports"
)

type createBillingUseCase struct {
	fxService services.FxService

	billingWorkflow ports.BillingWorkflow
}

type CreateBillingUsecase interface {
	Execute(ctx context.Context, userID string, description string, currency string, plannedClosedAt *time.Time) (string, error)
}

func NewCreateBillingUseCase(fxService services.FxService, billingWorkflow ports.BillingWorkflow) CreateBillingUsecase {
	return &createBillingUseCase{
		fxService:       fxService,
		billingWorkflow: billingWorkflow,
	}
}

func (uc *createBillingUseCase) Execute(ctx context.Context, userID string, description string, currency string, plannedClosedAt *time.Time) (string, error) {
	fn := "createBillingUseCase.CreateBilling"
	logger := rlog.With("fn", fn).With("userID", userID).With("description", description).With("currency", currency).With("plannedClosedAt", plannedClosedAt)

	// validate currency
	supportedCurrencies, err := uc.fxService.GetSupportedCurrencies(ctx, time.Now())
	if err != nil {
		logger.Error("failed to get supported currencies")
		return "", err
	}
	if !slices.Contains(supportedCurrencies, currency) {
		logger.Warn("currency not supported")
		return "", dto.ErrCurrencyNotSupported
	}

	// generate external billing ID
	randomUUID, err := uuid.NewV7()
	if err != nil {
		logger.Error("failed to generate external billing ID")
		return "", dto.ErrFailedToGenerateBillingID
	}
	externalBillingID := randomUUID.String()

	// get currency precision
	currencyMetadata, err := uc.fxService.GetCurrencyMetadata(ctx, currency, time.Now())
	if err != nil {
		logger.Error("failed to get currency metadata")
		return "", dto.ErrCurrencyMetadataNotFound
	}

	// start billing workflow
	err = uc.billingWorkflow.StartBilling(ctx, userID, externalBillingID, description, currency, currencyMetadata.Precision, plannedClosedAt)
	if err != nil {
		logger.Error("failed to start billing workflow")
		return "", dto.ErrFailedToStartBillingWorkflow
	}

	logger.Info("billing created successfully")

	// return external billing ID
	return externalBillingID, nil
}
