package usecases

import (
	"context"
	"slices"
	"time"

	"encore.dev/rlog"
	"github.com/google/uuid"

	"encore.app/billing/application/dto"
	"encore.app/billing/domain/repositories"
	"encore.app/billing/domain/services"
)

type createBillingUseCase struct {
	dbRepository repositories.DBRepository
	fxService    services.FxService
}

type CreateBillingUsecase interface {
	Execute(ctx context.Context, userID string, description string, currency string, plannedClosedAt *time.Time) (string, error)
}

func NewCreateBillingUseCase(dbRepository repositories.DBRepository, fxService services.FxService) CreateBillingUsecase {
	return &createBillingUseCase{
		dbRepository: dbRepository,
		fxService:    fxService,
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

	// create billing
	currencyPrecision := currencyMetadata.Precision
	externalBillingID, err = uc.dbRepository.CreateBilling(ctx, userID, externalBillingID, description, currency, currencyPrecision, plannedClosedAt)
	if err != nil {
		logger.Error("failed to create billing in database")
		return "", dto.ErrFailedToCreateBillingInDatabase
	}

	logger.Info("billing created successfully", "billingID", externalBillingID)

	// return external billing ID
	return externalBillingID, nil
}
