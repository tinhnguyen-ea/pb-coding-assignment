package persistence

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"encore.app/billing/domain/entities"
	"encore.app/billing/domain/repositories"
	"encore.dev/et"
)

func TestNewPostgresDBRepository(t *testing.T) {
	ctx := context.Background()
	db, _ := et.NewTestDatabase(ctx, "billing")

	repo := NewPostgresDBRepository(db)
	if repo == nil {
		t.Error("NewPostgresDBRepository returned nil")
	}

	// Verify it implements the interface
	var _ repositories.DBRepository = repo
}

func TestPostgresDBRepository_CreateBilling(t *testing.T) {
	ctx := context.Background()
	db, _ := et.NewTestDatabase(ctx, "billing")
	repo := NewPostgresDBRepository(db)
	externalBillingID, _ := uuid.NewV7()

	plannedClosedAt := time.Now().Add(24 * time.Hour)
	billingID, err := repo.CreateBilling(ctx, "user123", externalBillingID.String(), "Test billing", "USD", 2, &plannedClosedAt)
	if err != nil {
		t.Fatalf("CreateBilling failed: %v", err)
	}
	if billingID == 0 {
		t.Error("CreateBilling returned zero billing ID")
	}
}

func TestPostgresDBRepository_GetBillingByExternalID(t *testing.T) {
	ctx := context.Background()
	db, _ := et.NewTestDatabase(ctx, "billing")
	repo := NewPostgresDBRepository(db)
	externalBillingID, _ := uuid.NewV7()

	// Test not found
	_, err := repo.GetBillingByExternalID(ctx, externalBillingID.String())
	if err == nil {
		t.Error("Expected error for non-existent billing")
	}
	if !errors.Is(err, entities.ErrBillingNotFound) {
		t.Errorf("Expected ErrBillingNotFound, got: %v", err)
	}

	// Test found
	plannedClosedAt := time.Now().Add(24 * time.Hour)
	billingID, err := repo.CreateBilling(ctx, "user123", externalBillingID.String(), "Test billing", "USD", 2, &plannedClosedAt)
	if err != nil {
		t.Fatalf("CreateBilling failed: %v", err)
	}

	billing, err := repo.GetBillingByExternalID(ctx, externalBillingID.String())
	if err != nil {
		t.Fatalf("GetBillingByExternalID failed: %v", err)
	}
	if billing.ID != billingID {
		t.Errorf("Expected billing ID %d, got %d", billingID, billing.ID)
	}
	if billing.Status != entities.BillingStatusOpen {
		t.Errorf("Expected status %s, got %s", entities.BillingStatusOpen, billing.Status)
	}
}

func TestPostgresDBRepository_AddLineItem(t *testing.T) {
	ctx := context.Background()
	db, _ := et.NewTestDatabase(ctx, "billing")
	repo := NewPostgresDBRepository(db)
	externalBillingID, _ := uuid.NewV7()

	plannedClosedAt := time.Now().Add(24 * time.Hour)
	billingID, err := repo.CreateBilling(ctx, "user123", externalBillingID.String(), "Test billing", "USD", 2, &plannedClosedAt)
	if err != nil {
		t.Fatalf("CreateBilling failed: %v", err)
	}

	err = repo.AddLineItem(ctx, billingID, "Test item", 1000)
	if err != nil {
		t.Fatalf("AddLineItem failed: %v", err)
	}
}

func TestPostgresDBRepository_CloseBilling(t *testing.T) {
	ctx := context.Background()
	db, _ := et.NewTestDatabase(ctx, "billing")
	repo := NewPostgresDBRepository(db)
	externalBillingID, _ := uuid.NewV7()

	plannedClosedAt := time.Now().Add(24 * time.Hour)
	billingID, err := repo.CreateBilling(ctx, "user123", externalBillingID.String(), "Test billing", "USD", 2, &plannedClosedAt)
	if err != nil {
		t.Fatalf("CreateBilling failed: %v", err)
	}

	actualClosedAt := time.Now().UTC()
	err = repo.CloseBilling(ctx, billingID, actualClosedAt)
	if err != nil {
		t.Fatalf("CloseBilling failed: %v", err)
	}

	// Verify billing is closed
	billing, err := repo.GetBillingByExternalID(ctx, externalBillingID.String())
	if err != nil {
		t.Fatalf("GetBillingByExternalID failed: %v", err)
	}
	if billing.Status != entities.BillingStatusClosed {
		t.Errorf("Expected status %s, got %s", entities.BillingStatusClosed, billing.Status)
	}
}
