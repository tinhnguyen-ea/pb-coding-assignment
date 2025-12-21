package billing

import (
	"context"
	"errors"
	"time"

	"encore.dev"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
	"encore.dev/storage/sqldb"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"encore.app/billing/infrastructure/persistence"
	"encore.app/billing/infrastructure/services"
	"encore.app/billing/infrastructure/temporal"
	"encore.app/billing/infrastructure/temporal/activities"
	"encore.app/billing/infrastructure/temporal/workflows"
	"encore.app/billing/usecases"
	"encore.app/billing/usecases/dto"
)

var (
	// initialise database
	db = sqldb.NewDatabase("billing", sqldb.DatabaseConfig{
		Migrations: "./migrations",
	})

	// initialise workflow task queue
	envName                  = encore.Meta().Environment.Name
	billingWorkflowTaskQueue = envName + "-billing-workflow"
)

// encore:service
type Service struct {
	createBillingUsecase usecases.CreateBillingUsecase
	addLineItemUsecase   usecases.AddLineItemUsecase
	closeBillingUsecase  usecases.CloseBillingUsecase

	client client.Client
	worker worker.Worker
}

func initService() (*Service, error) {
	logger := rlog.With("fn", "initService")

	// initialise database repository
	dbRepository := persistence.NewPostgresDBRepository(db)

	// initialise FX service
	fxService := services.NewFxService()

	// initialise temporal client
	temporalClient, err := client.Dial(client.Options{})
	if err != nil {
		logger.Error("failed to dial temporal client", "error", err)
		return nil, err
	}

	// initialise billing workflow
	billingWorkflow := temporal.NewTemporalBillingWorkflow(temporalClient, billingWorkflowTaskQueue)

	// initialise create billing usecase
	createBillingUsecase := usecases.NewCreateBillingUseCase(fxService, billingWorkflow)

	// initialise add line item usecase
	addLineItemUsecase := usecases.NewAddLineItemUsecase(dbRepository, billingWorkflow)

	// initialise close billing usecase
	closeBillingUsecase := usecases.NewCloseBillingUseCase(dbRepository, billingWorkflow)

	// initialise temporal activities
	billingActivities := activities.NewBillingActivities(dbRepository, temporalClient, billingWorkflowTaskQueue)
	activities.SetActivityInstance(billingActivities)

	// initialise temporal worker
	temporalWorker := worker.New(temporalClient, billingWorkflowTaskQueue, worker.Options{})

	// register workflows
	temporalWorker.RegisterWorkflow(workflows.BillingWorkflow)

	// register activities
	temporalWorker.RegisterActivity(activities.StartBillingActivityFunc)
	temporalWorker.RegisterActivity(activities.AddLineItemActivityFunc)
	temporalWorker.RegisterActivity(activities.CloseBillingActivityFunc)
	temporalWorker.RegisterActivity(activities.CreateBillingSummaryActivityFunc)

	// start worker in background
	go func() {
		err := temporalWorker.Run(worker.InterruptCh())
		if err != nil {
			logger.Error("temporal worker stopped with error", "error", err)
		}
	}()

	logger.Info("Temporal worker started", "taskQueue", billingWorkflowTaskQueue)

	return &Service{
		createBillingUsecase: createBillingUsecase,
		addLineItemUsecase:   addLineItemUsecase,
		closeBillingUsecase:  closeBillingUsecase,

		client: temporalClient,
		worker: temporalWorker,
	}, nil
}

func (s *Service) Shutdown(ctx context.Context) {
	s.client.Close()
	s.worker.Stop()
}

// encore:api public method=POST path=/billing
func (s *Service) CreateBilling(ctx context.Context, req *CreateBillingRequest) (*CreateBillingResponse, error) {
	fn := "billing.Service.CreateBilling"
	logger := rlog.With("fn", fn).With("UserID", req.UserID)

	// validation user id
	if req.UserID == "" {
		logger.Warn("user ID is invalid")
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "user ID is required",
		}
	}

	// validate currency
	if req.Currency == "" {
		logger.Warn("currency is invalid")

		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "currency is required",
		}
	}

	// validate planned closed at
	if req.PlannedClosedAt != nil && req.PlannedClosedAt.Before(time.Now().UTC()) {
		logger.Warn("planned closed at is in the past")
		return nil, &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "planned closed at is in the past",
		}
	}

	logger.Info("Creating billing", "description", req.Description, "currency", req.Currency, "plannedClosedAt", req.PlannedClosedAt)
	billingID, err := s.createBillingUsecase.Execute(ctx, req.UserID, req.Description, req.Currency, req.PlannedClosedAt)
	if err != nil {
		if errors.Is(err, dto.ErrCurrencyNotSupported) {
			logger.Warn("currency not supported")
			return nil, &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "currency not supported",
			}
		}
		if errors.Is(err, dto.ErrFailedToGenerateBillingID) {
			logger.Warn("failed to generate billing ID")
			return nil, &errs.Error{
				Code:    errs.Internal,
				Message: "failed to generate billing ID",
			}
		}
		if errors.Is(err, dto.ErrCurrencyMetadataNotFound) {
			logger.Warn("currency metadata not found")
			return nil, &errs.Error{
				Code:    errs.Internal,
				Message: "currency metadata not found",
			}
		}
		if errors.Is(err, dto.ErrFailedToCreateBillingInDatabase) {
			logger.Warn("failed to create billing in database")
			return nil, &errs.Error{
				Code:    errs.Internal,
				Message: "failed to create billing in database",
			}
		}

		// unknown error
		logger.Error("Failed to create billing", "error", err)
		return nil, &errs.Error{
			Code:    errs.Internal,
			Message: "failed to create billing",
		}
	}

	logger.Info("Billing created successfully", "billingID", billingID)

	return &CreateBillingResponse{
		BillingID: billingID,
	}, nil
}

// encore:api public method=POST path=/billing/:billingID/line-item
func (s *Service) AddLineItem(ctx context.Context, billingID string, req *AddLineItemRequest) error {
	fn := "billing.Service.AddLineItem"
	logger := rlog.With("fn", fn).With("billingID", billingID).With("Amount", req.Amount)

	// validation amount
	if req.Amount <= 0 {
		logger.Warn("amount must be greater than 0")
		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "amount must be greater than 0",
		}
	}

	logger.Info("Adding line item to billing", "description", req.Description, "amount", req.Amount)

	err := s.addLineItemUsecase.Execute(ctx, billingID, req.Description, req.Amount)
	if err != nil {
		if errors.Is(err, dto.ErrBillingNotFound) {
			logger.Warn("billing not found")
			return &errs.Error{
				Code:    errs.NotFound,
				Message: "billing not found",
			}
		}
		if errors.Is(err, dto.ErrAmountHasTooManyDecimals) {
			logger.Warn("amount has too many decimals")
			return &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "amount has many decimals",
			}
		}
		if errors.Is(err, dto.ErrBillingNotOpen) {
			logger.Warn("billing is not open")
			return &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "billing is not open",
			}
		}

		logger.Error("failed to add line item", "error", err)
		// unknown error
		return &errs.Error{
			Code:    errs.Internal,
			Message: "failed to add line item",
		}
	}

	logger.Info("Line item added successfully", "billingID", billingID)

	return nil
}

// encore:api public method=POST path=/billing/:billingID/close
func (s *Service) CloseBilling(ctx context.Context, billingID string) error {
	fn := "billing.Service.CloseBilling"
	logger := rlog.With("fn", fn).With("billingID", billingID)

	// validation billing ID
	if billingID == "" {
		logger.Warn("billing ID is invalid")

		return &errs.Error{
			Code:    errs.InvalidArgument,
			Message: "billing ID is required",
		}
	}

	err := s.closeBillingUsecase.Execute(ctx, billingID)
	if err != nil {
		if errors.Is(err, dto.ErrBillingNotFound) {
			logger.Warn("billing not found")

			return &errs.Error{
				Code:    errs.NotFound,
				Message: "billing not found",
			}
		}
		if errors.Is(err, dto.ErrBillingNotOpen) {
			logger.Warn("billing is not open")

			return &errs.Error{
				Code:    errs.InvalidArgument,
				Message: "billing is not open",
			}
		}

		// unknown error
		logger.Error("failed to update billing", "error", err)
		return &errs.Error{
			Code:    errs.Internal,
			Message: "failed to update billing",
		}
	}

	logger.Info("Billing closed successfully", "billingID", billingID)

	return nil
}
