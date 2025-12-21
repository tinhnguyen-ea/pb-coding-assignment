package usecases

import (
	"context"
	"slices"
	"time"

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
	CreateBilling(ctx context.Context, userID string, description string, currency string, plannedClosedAt *time.Time) (string, error)
}

func NewCreateBillingUseCase(dbRepository repositories.DBRepository, fxService services.FxService) CreateBillingUsecase {
	return &createBillingUseCase{
		dbRepository: dbRepository,
		fxService:    fxService,
	}
}

func (uc *createBillingUseCase) CreateBilling(ctx context.Context, userID string, description string, currency string, plannedClosedAt *time.Time) (string, error) {
	// validate currency
	supportedCurrencies, err := uc.fxService.GetSupportedCurrencies(ctx, time.Now())
	if err != nil {
		return "", err
	}
	if !slices.Contains(supportedCurrencies, currency) {
		return "", dto.ErrCurrencyNotSupported
	}

	// generate external billing ID
	randomUUID, err := uuid.NewV7()
	if err != nil {
		return "", dto.ErrFailedToGenerateBillingID
	}
	externalBillingID := randomUUID.String()

	// get currency precision
	currencyMetadata, err := uc.fxService.GetCurrencyMetadata(ctx, currency, time.Now())
	if err != nil {
		return "", dto.ErrCurrencyMetadataNotFound
	}

	// create billing
	currencyPrecision := currencyMetadata.Precision
	externalBillingID, err = uc.dbRepository.CreateBilling(ctx, userID, externalBillingID, description, currency, currencyPrecision, plannedClosedAt)
	if err != nil {
		return "", dto.ErrFailedToCreateBillingInDatabase
	}

	// return external billing ID
	return externalBillingID, nil
}
