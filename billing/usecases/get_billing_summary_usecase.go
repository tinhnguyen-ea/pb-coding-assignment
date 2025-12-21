package usecases

import (
	"context"

	"encore.dev/beta/errs"
	"encore.dev/rlog"

	"encore.app/billing/domain/entities"
	"encore.app/billing/domain/repositories"
	"encore.app/billing/usecases/dto"
	"encore.app/billing/usecases/ports"
)

type GetBillingSummaryUseCase interface {
	Execute(ctx context.Context, billingID string) (*entities.BillingSummary, error)
}

type getBillingSummaryUseCase struct {
	dbRepository    repositories.DBRepository
	billingWorkflow ports.BillingWorkflow
}

func NewGetBillingSummaryUseCase(dbRepository repositories.DBRepository, billingWorkflow ports.BillingWorkflow) GetBillingSummaryUseCase {
	return &getBillingSummaryUseCase{
		dbRepository:    dbRepository,
		billingWorkflow: billingWorkflow,
	}
}

func (u *getBillingSummaryUseCase) Execute(ctx context.Context, billingID string) (*entities.BillingSummary, error) {
	fn := "usecases.getBillingSummaryUseCase.Execute"
	logger := rlog.With("fn", fn).With("billingID", billingID)

	// validate if billing exists
	billing, err := u.dbRepository.GetBillingByExternalID(ctx, billingID)
	if err != nil {
		logger.Error("failed to get billing by external ID", "error", err)
		return nil, err
	}
	if billing == nil {
		logger.Warn("billing not found")
		return nil, dto.ErrBillingNotFound
	}

	if billing.Status == entities.BillingStatusClosed {
		// get billing summary from database
		summary, err := u.dbRepository.GetBillingSummary(ctx, billingID)
		if err != nil {
			logger.Error("failed to get billing summary", "error", err)
			return nil, err
		}
		if summary == nil {
			logger.Warn("billing summary not found")
			return nil, &errs.Error{
				Code:    errs.NotFound,
				Message: "billing summary not found",
			}
		}

		return summary, nil
	} else {
		// get billing summary from temporal workflow
		summary, err := u.billingWorkflow.GetBillingSummary(ctx, billingID)
		if err != nil {
			logger.Error("failed to get billing summary", "error", err)
			return nil, err
		}
		return summary, nil
	}
}
