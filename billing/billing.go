package billing

import (
	"context"
	"errors"

	"encore.dev/beta/errs"
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

	return &Service{
		createBillingUsecase: createBillingUsecase,
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
