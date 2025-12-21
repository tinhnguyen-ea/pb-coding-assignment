package persistence

import (
	"context"
	"errors"
	"time"

	"encore.dev/rlog"
	"encore.dev/storage/sqldb"

	"encore.app/billing/domain/entities"
	"encore.app/billing/domain/repositories"
)

type postgresDBRepository struct {
	db *sqldb.Database
}

func NewPostgresDBRepository(db *sqldb.Database) repositories.DBRepository {
	return &postgresDBRepository{db: db}
}

func (r *postgresDBRepository) GetBillingByExternalID(ctx context.Context, externalBillingID string) (*entities.Billing, error) {
	fn := "infrastructure.persistence.postgresDBRepository.GetBillingByExternalID"
	logger := rlog.With("fn", fn).With("externalBillingID", externalBillingID)

	var billing entities.Billing

	// get billing from database
	err := r.db.QueryRow(ctx, `
		SELECT id, user_id, description, currency, currency_precision, status, planned_closed_at, actual_closed_at, created_at, updated_at FROM billings WHERE external_billing_id = $1
	`, externalBillingID).Scan(&billing.ID, &billing.UserID, &billing.Description, &billing.Currency, &billing.CurrencyPrecision, &billing.Status, &billing.PlannedClosedAt, &billing.ActualClosedAt, &billing.CreatedAt, &billing.UpdatedAt)
	if err != nil {
		// no rows found
		if errors.Is(err, sqldb.ErrNoRows) {
			logger.Warn("Billing not found")
			return nil, entities.ErrBillingNotFound
		}

		// unknown error
		logger.Error("Failed to get billing by external ID", "error", err)
		return nil, entities.ErrDBService
	}

	return &billing, nil
}

func (r *postgresDBRepository) CreateBilling(ctx context.Context, userID string, externalBillingID string, description string, currency string, currencyPrecision int64, plannedClosedAt *time.Time) (int64, error) {
	fn := "infrastructure.persistence.postgresDBRepository.CreateBilling"
	logger := rlog.With("fn", fn).With("userID", userID).With("externalBillingID", externalBillingID).With("description", description).With("currency", currency).With("currencyPrecision", currencyPrecision).With("plannedClosedAt", plannedClosedAt)

	var billingID int64

	// insert billing into database
	err := r.db.QueryRow(ctx, `
		INSERT INTO billings (user_id, external_billing_id, description, currency, currency_precision, status, planned_closed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, userID, externalBillingID, description, currency, currencyPrecision, entities.BillingStatusOpen, plannedClosedAt).Scan(&billingID)
	if err != nil {
		logger.Error("failed to create billing in database", "error", err)
		return 0, entities.ErrDBService
	}

	logger.Info("billing created successfully")

	// return billing ID
	return billingID, nil
}

func (r *postgresDBRepository) AddLineItem(ctx context.Context, billingID int64, description string, amountMinor int64) error {
	fn := "infrastructure.persistence.postgresDBRepository.AddLineItem"
	logger := rlog.With("fn", fn).With("billingID", billingID).With("description", description).With("amountMinor", amountMinor)

	// insert line item into database
	_, err := r.db.Exec(ctx, `
		INSERT INTO line_items (billing_id, description, amount_minor)
		VALUES ($1, $2, $3)
	`, billingID, description, amountMinor)
	if err != nil {
		logger.Error("failed to add line item to database", "error", err)
		return entities.ErrDBService
	}

	logger.Info("line item added successfully")

	return nil
}

func (r *postgresDBRepository) CloseBilling(ctx context.Context, billingID int64, actualClosedAt time.Time) error {
	fn := "infrastructure.persistence.postgresDBRepository.CloseBilling"
	logger := rlog.With("fn", fn).With("billingID", billingID).With("actualClosedAt", actualClosedAt)

	// update billing in database
	_, err := r.db.Exec(ctx, `
		UPDATE billings SET status = $1, actual_closed_at = $2 WHERE id = $3
	`, entities.BillingStatusClosed, actualClosedAt, billingID)
	if err != nil {
		logger.Error("failed to close billing in database", "error", err)
		return entities.ErrDBService
	}

	logger.Info("billing closed successfully")

	return nil
}

func (r *postgresDBRepository) CreateBillingSummary(ctx context.Context, externalBillingID string, billingSummary []byte) error {
	fn := "infrastructure.persistence.postgresDBRepository.CreateBillingSummary"
	logger := rlog.With("fn", fn).With("externalBillingID", externalBillingID)

	// insert billing summary into database
	_, err := r.db.Exec(ctx, `
		INSERT INTO billing_summaries (external_billing_id, summary)
		VALUES ($1, $2)
	`, externalBillingID, billingSummary)
	if err != nil {
		logger.Error("failed to create billing summary in database", "error", err)
		return entities.ErrDBService
	}

	logger.Info("billing summary created successfully")

	return nil
}

func (r *postgresDBRepository) GetBillingSummary(ctx context.Context, externalBillingID string) (*entities.BillingSummary, error) {
	fn := "infrastructure.persistence.postgresDBRepository.GetBillingSummary"
	logger := rlog.With("fn", fn).With("externalBillingID", externalBillingID)

	// get billing summary from database
	var summary entities.BillingSummary
	err := r.db.QueryRow(ctx, `
		SELECT summary FROM billing_summaries WHERE external_billing_id = $1
	`, externalBillingID).Scan(&summary)
	if err != nil {
		logger.Error("failed to get billing summary from database", "error", err)
		return nil, entities.ErrDBService
	}

	return &summary, nil
}
