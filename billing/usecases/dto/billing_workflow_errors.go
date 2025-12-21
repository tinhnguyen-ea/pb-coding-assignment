package dto

import "errors"

var (
	ErrFailedToStartBillingWorkflow         = errors.New("failed to start billing workflow")
	ErrFailedToAddLineItemToWorkflow        = errors.New("failed to add line item to workflow")
	ErrFailedToCloseBillingWorkflow         = errors.New("failed to close billing workflow")
	ErrFailedToAddLineItemToBillingWorkflow = errors.New("failed to add line item to billing workflow")
	ErrFailedToCloseBillingInWorkflow       = errors.New("failed to close billing in workflow")
)
