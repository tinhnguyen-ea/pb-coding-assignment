package usecases

import (
	"context"
	"errors"

	"encore.app/billing/domain/entities"
	"encore.app/billing/domain/repositories"
	"encore.app/billing/usecases/dto"
	"encore.app/billing/usecases/ports"
	"encore.dev/rlog"
)

type closeBillingUseCase struct {
	dbRepository    repositories.DBRepository
	billingWorkflow ports.BillingWorkflow
}

type CloseBillingUsecase interface {
	Execute(ctx context.Context, externalBillingID string) error
}

func NewCloseBillingUseCase(dbRepository repositories.DBRepository, billingWorkflow ports.BillingWorkflow) CloseBillingUsecase {
	return &closeBillingUseCase{dbRepository: dbRepository, billingWorkflow: billingWorkflow}
}

func (uc *closeBillingUseCase) Execute(ctx context.Context, externalBillingID string) error {
	fn := "closeBillingUseCase.CloseBilling"
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
	err = uc.billingWorkflow.CloseBilling(ctx, externalBillingID)
	if err != nil {
		logger.Error("failed to close billing", "error", err)
		return dto.ErrFailedToCloseBillingInWorkflow
	}

	logger.Info("billing closed successfully", "billingID", billing.ID)

	return nil
}
