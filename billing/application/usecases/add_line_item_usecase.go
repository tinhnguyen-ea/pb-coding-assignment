package usecases

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"

	"encore.dev/rlog"

	"encore.app/billing/application/dto"
	"encore.app/billing/domain/entities"
	"encore.app/billing/domain/repositories"
)

type addLineItemUseCase struct {
	dbRepository repositories.DBRepository
}

type AddLineItemUsecase interface {
	AddLineItem(ctx context.Context, externalBillingID string, description string, amount float64) error
}

func NewAddLineItemUsecase(dbRepository repositories.DBRepository) AddLineItemUsecase {
	return &addLineItemUseCase{dbRepository: dbRepository}
}

func (uc *addLineItemUseCase) AddLineItem(ctx context.Context, externalBillingID string, description string, amount float64) error {
	logger := rlog.With("externalBillingID", externalBillingID).With("amount", amount)

	// get billing
	billing, err := uc.dbRepository.GetBillingByExternalID(ctx, externalBillingID)
	if err != nil {
		if errors.Is(err, entities.ErrBillingNotFound) {
			logger.Warn("Billing not found")
			return dto.ErrBillingNotFound
		}

		// unknown error
		logger.Error("Failed to get billing by external ID", "error", err)
		return dto.ErrFailedToGetBillingByExternalID
	}

	// validate billing is open
	if billing.Status != entities.BillingStatusOpen {
		logger.Warn("billing is not open")
		return dto.ErrBillingNotOpen
	}

	if !hasAtMostXDecimals(amount, billing.CurrencyPrecision) {
		logger.Warn("Amount has many decimals")
		return dto.ErrAmountHasManyDecimals
	}

	// convert amount to minor units
	currencyPrecision := billing.CurrencyPrecision
	amountMinor := int64(amount * math.Pow10(int(currencyPrecision)))

	// add line item
	err = uc.dbRepository.AddLineItem(ctx, billing.ID, description, amountMinor)
	if err != nil {
		logger.Error("Failed to add line item to database", "error", err)
		return dto.ErrFailedToAddLineItemToDatabase
	}

	logger.Info("Line item added successfully")

	return nil
}

func hasAtMostXDecimals(f float64, x int64) bool {
	s := fmt.Sprintf(fmt.Sprintf("%%.%df", x), f)
	converted, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return false
	}

	epsilon := 1e-9
	return math.Abs(f-converted) < epsilon
}
