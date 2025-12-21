package usecases

import (
	"context"
	"errors"
	"math"

	"encore.dev/rlog"

	"encore.app/billing/domain/entities"
	"encore.app/billing/domain/repositories"
	"encore.app/billing/usecases/dto"
	"encore.app/billing/usecases/ports"
)

type addLineItemUseCase struct {
	dbRepository    repositories.DBRepository
	billingWorkflow ports.BillingWorkflow
}

type AddLineItemUsecase interface {
	Execute(ctx context.Context, externalBillingID string, description string, amount float64) error
}

func NewAddLineItemUsecase(dbRepository repositories.DBRepository, billingWorkflow ports.BillingWorkflow) AddLineItemUsecase {
	return &addLineItemUseCase{dbRepository: dbRepository, billingWorkflow: billingWorkflow}
}

func (uc *addLineItemUseCase) Execute(ctx context.Context, externalBillingID string, description string, amount float64) error {
	fn := "addLineItemUseCase.AddLineItem"
	logger := rlog.With("fn", fn).With("externalBillingID", externalBillingID).With("amount", amount)

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

	// validate billing is open
	if !billing.CanAddLineItem() {
		logger.Warn("billing is not open")
		return dto.ErrBillingNotOpen
	}

	if !billing.CanAddItemWithAmount(amount) {
		logger.Warn("amount has too many decimals")
		return dto.ErrAmountHasTooManyDecimals
	}

	// convert amount to minor units
	currencyPrecision := billing.CurrencyPrecision
	amountMinor := int64(amount * math.Pow10(int(currencyPrecision)))

	// add line item to billing workflow
	err = uc.billingWorkflow.AddLineItem(ctx, externalBillingID, description, amountMinor)
	if err != nil {
		logger.Error("failed to add line item to billing workflow", "error", err)
		return dto.ErrFailedToAddLineItemToBillingWorkflow
	}

	logger.Info("line item added successfully")

	return nil
}
