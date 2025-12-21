package usecases

import (
	"context"
	"errors"
	"time"

	"encore.app/billing/application/dto"
	"encore.app/billing/domain/entities"
	"encore.app/billing/domain/repositories"
)

type updateBillingUseCase struct {
	dbRepository repositories.DBRepository
}

type UpdateBillingUsecase interface {
	CloseBilling(ctx context.Context, externalBillingID string) error
}

func NewUpdateBillingUseCase(dbRepository repositories.DBRepository) UpdateBillingUsecase {
	return &updateBillingUseCase{dbRepository: dbRepository}
}

func (uc *updateBillingUseCase) CloseBilling(ctx context.Context, externalBillingID string) error {
	// get billing
	billing, err := uc.dbRepository.GetBillingByExternalID(ctx, externalBillingID)
	if err != nil {
		if errors.Is(err, entities.ErrBillingNotFound) {
			return dto.ErrBillingNotFound
		}

		// unknown error
		return dto.ErrFailedToGetBillingByExternalID
	}
	if billing.Status != entities.BillingStatusOpen {
		return dto.ErrBillingNotOpen
	}

	// close billing
	err = uc.dbRepository.CloseBilling(ctx, billing.ID, time.Now().UTC())
	if err != nil {
		return dto.ErrFailedToCloseBillingInDatabase
	}

	return nil
}
