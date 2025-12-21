package dto

import "errors"

var (
	ErrCurrencyNotSupported            = errors.New("currency not supported")
	ErrCurrencyMetadataNotFound        = errors.New("currency metadata not found in FX service")
	ErrFailedToCreateBillingInDatabase = errors.New("failed to create billing in database")
	ErrFailedToGenerateBillingID       = errors.New("failed to generate billing ID")
)
