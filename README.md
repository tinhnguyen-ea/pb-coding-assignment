# Billing Service

A billing service built with Encore.go, implementing Clean Architecture principles and using Temporal for workflow orchestration.

## Table of Contents

- [Overview](#overview)
- [Assumption](#assumption)
- [Architecture](#architecture)
- [Database Design](#database-design)
- [Technology Stack](#technology-stack)
- [Project Structure](#project-structure)
- [API Endpoints](#api-endpoints)
- [Workflow Orchestration](#workflow-orchestration)
- [Getting Started](#getting-started)

## Overview

This billing service provides a solution for managing billing operations, including:

- Creating billings with multi-currency support
- Adding line items to open billings
- Closing billings (manually or automatically)
- Generating billing summaries

The service follows Clean Architecture principles, ensuring separation of concerns, testability, and maintainability. It leverages Temporal workflows for reliable, long-running orchestration of billing operations.

## Assumption
- The system supports two billing models: period billing and open-ended billing. Period billing includes a predefined close time, whereas open-ended billing has no fixed close time and must be explicitly closed through an API endpoint.
- Authentication and authorization are handled by an external service or API gateway. The billing service is deployed within a private network and is not publicly accessible; all of its APIs are internal-only.
- The `fx` service is intentionally overcomplicated within the scope of this assignment to showcase Encore.dev’s tracing features and the billing service's architecture.
- The billing summary is intentionally minimal and focuses on demonstrating Temporal features.

## Architecture

The project follows **Clean Architecture** with clear separation between layers:

### Architecture Layers

```
┌─────────────────────────────────────────────────────────┐
│                    API Layer (billing.go)               │
│              HTTP endpoints, validation,                │
│              error handling, request/response           │
└──────────────────────┬──────────────────────────────────┘
                       │
┌──────────────────────▼──────────────────────────────────┐
│              Use Cases Layer (usecases/)                │
│     Business logic orchestration, application rules     │
└──────────────┬────────────────────┬─────────────────────┘
               │                    │
┌──────────────▼──────────┐ ┌──────▼──────────────────────┐
│   Domain Layer          │ │  Infrastructure Layer       │
│   - Entities            │ │  - Persistence (DB)         │
│   - Repositories        │ │  - Services (FX)            │
│   - Services            │ │  - Temporal (Workflows)     │
└─────────────────────────┘ └─────────────────────────────┘
```

### Layer Responsibilities

1. **API Layer** (`billing/billing.go`)
   - HTTP request/response handling
   - Input validation
   - Error translation to HTTP status codes
   - Encore API annotations

2. **Use Cases Layer** (`billing/usecases/`)
   - Business logic orchestration
   - Coordinates between domain entities and infrastructure
   - Handles application-specific rules and workflows
   - Use cases:
     - `CreateBillingUsecase`: Creates new billing with currency validation
     - `AddLineItemUsecase`: Adds line items to open billings
     - `CloseBillingUsecase`: Closes billings and triggers summary generation
     - `GetBillingSummaryUsecase`: Retrieves billing summaries

3. **Domain Layer** (`billing/domain/`)
   - **Entities** (`entities/`): Core business objects (Billing, LineItem, BillingSummary)
   - **Repositories** (`repositories/`): Interfaces for data access
   - **Services** (`services/`): Domain service interfaces (e.g., FxService)
   - Business rules and validation logic

4. **Infrastructure Layer** (`billing/infrastructure/`)
   - **Persistence** (`persistence/`): PostgreSQL repository implementation
   - **Services** (`services/`): External service implementations (FX service)
   - **Temporal** (`temporal/`): Workflow orchestration
     - Workflows: Long-running billing state management
     - Activities: Atomic operations (database operations)

### Design Patterns

- **Dependency Inversion**: Domain layer defines interfaces; infrastructure implements them
- **Repository Pattern**: Data access abstraction
- **Use Case Pattern**: Encapsulates business workflows
- **Workflow Orchestration**: Temporal workflows manage long-running billing processes

## Database Design

The service uses PostgreSQL with the following schema:

### Tables

#### `billings`
Stores billing records with currency and status information.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key (internal ID) |
| `external_billing_id` | UUID | Public-facing billing identifier (unique) |
| `user_id` | TEXT | User identifier |
| `description` | TEXT | Billing description |
| `currency` | CURRENCY_CODE | Currency code (enum) |
| `currency_precision` | SMALLINT | Decimal places for currency |
| `status` | BILLING_STATUS | Current status: 'open' or 'closed' |
| `planned_closed_at` | TIMESTAMPTZ | Scheduled auto-close time (nullable) |
| `actual_closed_at` | TIMESTAMPTZ | Actual close time (nullable) |
| `created_at` | TIMESTAMPTZ | Record creation timestamp |
| `updated_at` | TIMESTAMPTZ | Last update timestamp |

**Indexes:**
- `billing_user_id_with_status_idx`: Composite index on `(user_id, status)` for querying user billings
- `billing_external_billing_id_idx`: Hash index on `external_billing_id` for fast lookups

#### `line_items`
Stores individual line items associated with billings.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `billing_id` | BIGINT | Foreign key to `billings.id` |
| `description` | TEXT | Line item description |
| `amount_minor` | BIGINT | Amount in minor currency units (e.g., cents) |
| `created_at` | TIMESTAMPTZ | Record creation timestamp |
| `updated_at` | TIMESTAMPTZ | Last update timestamp |

**Foreign Key:**
- `billing_id` → `billings.id` (cascade on delete)

#### `billing_summaries`
Stores generated billing summaries as JSONB for fast retrieval.

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `external_billing_id` | UUID | Foreign key to billing (unique) |
| `summary` | JSONB | Complete billing summary data |
| `created_at` | TIMESTAMPTZ | Record creation timestamp |
| `updated_at` | TIMESTAMPTZ | Last update timestamp |

**Index:**
- Unique constraint on `external_billing_id`

### Enums

#### `BILLING_STATUS`
- `'open'`: Billing is active and can accept line items
- `'closed'`: Billing is finalized and cannot be modified

#### `CURRENCY_CODE`
Supports 2 currencies: USD, GEL

### Design Decisions

1. **Dual ID System**: Internal `id` (BIGSERIAL) for database efficiency, `external_billing_id` (UUID) for public API
2. **Amount Storage**: All amounts stored in minor units (e.g., cents) as `BIGINT` to avoid floating-point precision issues
3. **JSONB Summaries**: Billing summaries stored as JSONB for flexible schema and fast retrieval
4. **Indexing Strategy**: Optimized indexes for common query patterns (user lookups, external ID lookups)

## Technology Stack

- **Framework**: Encore.go v1.52.1
- **Language**: Go 1.25.4
- **Database**: PostgreSQL (via Encore SQL databases)
- **Workflow Engine**: Temporal SDK v1.38.0
- **Architecture**: Clean Architecture / Hexagonal Architecture

### Key Dependencies

- `encore.dev`: Encore framework core
- `go.temporal.io/sdk`: Temporal workflow orchestration
- `github.com/google/uuid`: UUID generation
- `github.com/jackc/pgx/v5`: PostgreSQL driver

## Project Structure

```
billing/
├── billing.go                          # Service entry point, API handlers
├── types.go                            # Request/response DTOs
├── migrations/                         # Database migrations
│   ├── 1_create_billing_tables.up.sql
│   └── 2_create_billing_summary.up.sql
├── domain/                             # Domain layer (business logic)
│   ├── entities/                       # Core business entities
│   │   ├── billing.go                  # Billing, LineItem, BillingSummary
│   │   ├── fx.go                       # Currency entities
│   │   └── errors.go                   # Domain errors
│   ├── repositories/                   # Repository interfaces
│   │   └── billing_repository.go
│   └── services/                       # Domain service interfaces
│       └── fx.go                       # FX service interface
├── usecases/                           # Application use cases
│   ├── create_billing_usecase.go
│   ├── add_line_item_usecase.go
│   ├── close_billing_usecase.go
│   ├── get_billing_summary_usecase.go
│   └── dto/                            # Use case DTOs and errors
├── infrastructure/                     # Infrastructure implementations
│   ├── persistence/                    # Database repository
│   │   └── db_billing.go
│   ├── services/                       # External service adapters
│   │   └── fx.go                       # FX service implementation
│   └── temporal/                       # Temporal workflow orchestration
│       ├── billing_workflow.go         # Workflow client wrapper
│       ├── workflows/                  # Workflow definitions
│       │   └── billing_workflow_definition.go
│       └── activities/                 # Activity implementations
│           └── billing_activities.go
└── fx/                                 # External FX service
    └── fx.go                           # Currency metadata and rates API
```

## API Endpoints

All endpoints are public and require proper authentication in production.

### POST `/billing`
Creates a new billing record.

**Request:**
```json
{
  "user_id": "user123",
  "description": "Monthly subscription",
  "currency": "USD",
  "planned_closed_at": "2024-12-31T23:59:59Z"  // optional
}
```

**Response:**
```json
{
  "billing_id": "550e8400-e29b-41d4-a716-446655440000"
}
```

### POST `/billing/:billingID/line-item`
Adds a line item to an open billing.

**Request:**
```json
{
  "description": "Premium feature",
  "amount": 29.99
}
```

**Response:** `204 No Content` on success

### POST `/billing/:billingID/close`
Manually closes a billing and triggers summary generation.

**Response:** `204 No Content` on success

### GET `/billing/:billingID/summary`
Retrieves the billing summary.

**Response:**
```json
{
  "billing_id": "550e8400-e29b-41d4-a716-446655440000",
  "description": "Monthly subscription",
  "currency": "USD",
  "currency_precision": 2,
  "line_items": [
    {
      "description": "Premium feature",
      "amountMinor": 2999
    }
  ],
  "total_amount_minor": 2999
}
```

## Workflow Orchestration

The service uses **Temporal** for reliable, long-running workflow orchestration of billing operations.

### Workflow Lifecycle

1. **Start**: Workflow starts when billing is created
   - Validates currency
   - Creates billing record in database
   - Initializes workflow state

2. **Active State**: Workflow waits for events
   - Listens for `add-line-item` signals
   - Listens for `close-billing` signals
   - Monitors auto-close timer (if `planned_closed_at` is set)

3. **Close**: Workflow closes billing
   - Updates billing status to 'closed'
   - Generates billing summary
   - Stores summary in database

### Workflow Components

#### Activities (Atomic Operations)
- `StartBillingActivity`: Creates billing in database
- `AddLineItemActivity`: Adds line item to database
- `CloseBillingActivity`: Closes billing in database
- `CreateBillingSummaryActivity`: Stores billing summary

#### Signals (Events)
- `add-line-item`: Triggers line item addition
- `close-billing`: Triggers manual billing closure

#### Query
- `currentState`: Returns current workflow state

### Benefits

- **Reliability**: Automatic retries and fault tolerance
- **Observability**: Complete workflow execution history
- **Scalability**: Handles concurrent billing operations
- **Consistency**: Ensures billing state consistency

## Getting Started

### Prerequisites

- Go 1.25.4 or later
- Encore CLI installed
- Docker (for local database and Temporal)

### Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd coding-challenge
   ```

2. **Install dependencies**
   ```bash
   go mod download
   ```

3. **Run locally**
   ```bash
   encore run
   ```

   This will:
   - Start the Encore development server
   - Provision a local PostgreSQL database
   - Start Temporal server (if configured)
   - Run database migrations automatically

4. **Run tests**
   ```bash
   encore test ./...
   ```

### Configuration

- Database migrations are automatically applied on startup
- Temporal connection uses default local configuration
- FX service is configured as a separate Encore service

### Environment

The service supports multiple environments:
- **Local**: Development environment with local database
- **Development**: Cloud development environment
- **Production**: Production environment with managed infrastructure

## Key Features

- ✅ Multi-currency support with precision handling
- ✅ Reliable workflow orchestration with Temporal
- ✅ Clean Architecture for maintainability
- ✅ Comprehensive error handling
- ✅ Database migrations
- ✅ Type-safe API contracts
- ✅ Automatic retry logic for workflows
- ✅ Auto-close billings based on scheduled time

## Notes

- All amounts are stored in minor units (e.g., cents) to avoid floating-point precision issues
- Billing status transitions are managed by workflows to ensure consistency
- External billing IDs (UUIDs) are used in public APIs; internal IDs are database-specific
- Billing summaries are generated asynchronously when billings are closed
