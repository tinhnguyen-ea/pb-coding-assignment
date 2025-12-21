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

func (r *postgresDBRepository) CreateBilling(ctx context.Context, userID string, externalBillingID string, description string, currency string, currencyPrecision int64, plannedClosedAt *time.Time) (string, error) {
	// insert billing into database
	_, err := r.db.Exec(ctx, `
		INSERT INTO billings (user_id, external_billing_id, description, currency, currency_precision, status, planned_closed_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`, userID, externalBillingID, description, currency, currencyPrecision, entities.BillingStatusOpen, plannedClosedAt)
	if err != nil {
		return "", entities.ErrDBService
	}

	// return external billing ID
	return externalBillingID, nil
}

func (r *postgresDBRepository) AddLineItem(ctx context.Context, externalBillingID string, description string, amountMinor int64) error {
	// insert line item into database
	_, err := r.db.Exec(ctx, `
		INSERT INTO line_items (billing_id, description, amount_minor)
		VALUES ($1, $2, $3)
	`, externalBillingID, description, amountMinor)
	if err != nil {
		return entities.ErrDBService
	}

	return nil
}

func (r *postgresDBRepository) CloseBilling(ctx context.Context, externalBillingID string, actualClosedAt time.Time) error {
	// update billing in database
	_, err := r.db.Exec(ctx, `
		UPDATE billings SET status = $1, actual_closed_at = $2 WHERE id = $3
	`, entities.BillingStatusClosed, actualClosedAt, externalBillingID)
	if err != nil {
		return entities.ErrDBService
	}

	return nil
}
