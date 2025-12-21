package persistence

import (
	"context"
	"time"

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
