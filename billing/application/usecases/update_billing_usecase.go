package usecases

import (
	"context"
	"errors"
	"time"

	"encore.app/billing/application/dto"
	"encore.app/billing/domain/entities"
	"encore.app/billing/domain/repositories"
	"encore.dev/rlog"
)

type updateBillingUseCase struct {
	dbRepository repositories.DBRepository
}

type UpdateBillingUsecase interface {
	Execute(ctx context.Context, externalBillingID string) error
}

func NewUpdateBillingUseCase(dbRepository repositories.DBRepository) UpdateBillingUsecase {
	return &updateBillingUseCase{dbRepository: dbRepository}
}

func (uc *updateBillingUseCase) Execute(ctx context.Context, externalBillingID string) error {
	fn := "updateBillingUseCase.CloseBilling"
	logger := rlog.With("fn", fn).With("externalBillingID", externalBillingID)

	// get billing
	billing, err := uc.dbRepository.GetBillingByExternalID(ctx, externalBillingID)
	if err != nil {
		if errors.Is(err, entities.ErrBillingNotFound) {
			logger.Warn("billing not found")
			return dto.ErrBillingNotFound
		}

		// unknown error
		logger.Error("failed to get billing by external ID", "error", err)
		return dto.ErrFailedToGetBillingByExternalID
	}
	if billing.Status != entities.BillingStatusOpen {
		logger.Warn("billing is not open")
		return dto.ErrBillingNotOpen
	}

	// close billing
	err = uc.dbRepository.CloseBilling(ctx, billing.ID, time.Now().UTC())
	if err != nil {
		logger.Error("failed to close billing in database", "error", err)
		return dto.ErrFailedToCloseBillingInDatabase
	}

	logger.Info("billing closed successfully", "billingID", billing.ID)

	return nil
}
