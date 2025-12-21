package usecases

import (
	"context"
	"errors"
	"math"

	"encore.dev/rlog"

	"encore.app/billing/application/dto"
	"encore.app/billing/domain/entities"
	"encore.app/billing/domain/repositories"
)

type addLineItemUseCase struct {
	dbRepository repositories.DBRepository
}

type AddLineItemUsecase interface {
	Execute(ctx context.Context, externalBillingID string, description string, amount float64) error
}

func NewAddLineItemUsecase(dbRepository repositories.DBRepository) AddLineItemUsecase {
	return &addLineItemUseCase{dbRepository: dbRepository}
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
		logger.Warn("amount has many decimals")
		return dto.ErrAmountHasManyDecimals
	}

	// convert amount to minor units
	currencyPrecision := billing.CurrencyPrecision
	amountMinor := int64(amount * math.Pow10(int(currencyPrecision)))

	// add line item
	err = uc.dbRepository.AddLineItem(ctx, billing.ID, description, amountMinor)
	if err != nil {
		logger.Error("failed to add line item to database", "error", err)
		return dto.ErrFailedToAddLineItemToDatabase
	}

	logger.Info("line item added successfully")

	return nil
}
