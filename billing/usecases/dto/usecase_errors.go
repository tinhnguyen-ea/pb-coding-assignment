package dto

import "errors"

var (
	ErrCurrencyNotSupported            = errors.New("currency not supported")
	ErrCurrencyMetadataNotFound        = errors.New("currency metadata not found in FX service")
	ErrFailedToCreateBillingInDatabase = errors.New("failed to create billing in database")
	ErrFailedToGenerateBillingID       = errors.New("failed to generate billing ID")

	ErrBillingNotFound                = errors.New("billing not found")
	ErrAmountHasTooManyDecimals       = errors.New("amount has too many decimals")
	ErrBillingNotOpen                 = errors.New("billing is not open")
	ErrFailedToGetBillingByExternalID = errors.New("failed to get billing by external ID")
	ErrFailedToAddLineItemToDatabase  = errors.New("failed to add line item to database")

	ErrFailedToCloseBillingInDatabase = errors.New("failed to close billing in database")
)
