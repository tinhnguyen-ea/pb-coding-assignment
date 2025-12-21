package billing

import (
	"context"
	"errors"

	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"

	"encore.app/billing/application/dto"
	"encore.app/billing/application/usecases"
	"encore.app/billing/infrastructure/persistence"
	"encore.app/billing/infrastructure/services"
)

// initialise database
var db = sqldb.NewDatabase("billing", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

// encore:service
type Service struct {
	createBillingUsecase usecases.CreateBillingUsecase
	addLineItemUsecase   usecases.AddLineItemUsecase
	updateBillingUsecase usecases.UpdateBillingUsecase
}

func initService() (*Service, error) {
	// initialise database repository
	dbRepository := persistence.NewPostgresDBRepository(db)

	// initialise FX service
	fxService := services.NewFxService()

	// initialise create billing usecase
	createBillingUsecase := usecases.NewCreateBillingUseCase(
		dbRepository,
		fxService,
	)

	// initialise add line item usecase
	addLineItemUsecase := usecases.NewAddLineItemUsecase(
		dbRepository,
	)

	// initialise update billing usecase
	updateBillingUsecase := usecases.NewUpdateBillingUseCase(
		dbRepository,
	)

	return &Service{
		createBillingUsecase: createBillingUsecase,
		addLineItemUsecase:   addLineItemUsecase,
		updateBillingUsecase: updateBillingUsecase,
	}, nil
}

// encore:api public method=POST path=/billing
func (s *Service) CreateBilling(ctx context.Context, req *CreateBillingRequest) (*CreateBillingResponse, error) {
	// validation user id
	if req.UserID == "" {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "user ID is required",
		}
	}

	// validate currency
	if req.Currency == "" {
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "currency is required",
		}
	}

	billingID, err := s.createBillingUsecase.CreateBilling(ctx, req.UserID, req.Description, req.Currency, req.PlannedClosedAt)
	if err != nil {
		if errors.Is(err, dto.ErrCurrencyNotSupported) {
			return nil, &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "currency not supported",
			}
		}
		if errors.Is(err, dto.ErrFailedToGenerateBillingID) {
			return nil, &errs.Error{
				Code:    errs.Internal,
				Message: "failed to generate billing ID",
			}
		}
		if errors.Is(err, dto.ErrCurrencyMetadataNotFound) {
			return nil, &errs.Error{
				Code:    errs.Internal,
				Message: "currency metadata not found",
			}
		}
		if errors.Is(err, dto.ErrFailedToCreateBillingInDatabase) {
			return nil, &errs.Error{
				Code:    errs.Internal,
				Message: "failed to create billing in database",
			}
		}

		// unknown error
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to create billing",
		}
	}

	return &CreateBillingResponse{
		BillingID: billingID,
	}, nil
}

// encore:api public method=POST path=/billing/:billingID/line-item
func (s *Service) AddLineItem(ctx context.Context, billingID string, req *AddLineItemRequest) error {
	logger := rlog.With("billingID", billingID).With("Amount", req.Amount)
	logger.Info("Adding line item to billing", "description", req.Description)

	// validation amount
	if req.Amount <= 0 {
		logger.Warn("Amount must be greater than 0")
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "amount must be greater than 0",
		}
	}

	err := s.addLineItemUsecase.AddLineItem(ctx, billingID, req.Description, req.Amount)
	if err != nil {
		if errors.Is(err, dto.ErrBillingNotFound) {
			logger.Warn("Billing not found")
			return &errs.Error{
				Code:    errs.NotFound,
				Message: "billing not found",
			}
		}
		if errors.Is(err, dto.ErrAmountHasManyDecimals) {
			logger.Warn("Amount has many decimals")
			return &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "amount has many decimals",
			}
		}
		if errors.Is(err, dto.ErrBillingNotOpen) {
			logger.Warn("Billing is not open")
			return &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "billing is not open",
			}
		}

		logger.Error("Failed to add line item", "error", err)
		// unknown error
		return &errs.Error{
			Code:    errs.Internal,
			Message: "failed to add line item",
		}
	}

	return nil
}

// encore:api public method=PATCH path=/billing/:billingID
func (s *Service) CloseBilling(ctx context.Context, billingID string) error {
	// validation billing ID
	if billingID == "" {
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "billing ID is required",
		}
	}

	err := s.updateBillingUsecase.CloseBilling(ctx, billingID)
	if err != nil {
		if errors.Is(err, dto.ErrBillingNotFound) {
			return &errs.Error{
				Code:    errs.NotFound,
				Message: "billing not found",
			}
		}
		if errors.Is(err, dto.ErrBillingNotOpen) {
			return &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "billing is not open",
			}
		}

		// unknown error
		return &errs.Error{
			Code:    errs.Internal,
			Message: "failed to update billing",
		}
	}

	return nil
}
