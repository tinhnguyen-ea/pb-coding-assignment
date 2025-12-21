CREATE TYPE BILLING_STATUS AS ENUM ('open', 'closed');
CREATE TYPE CURRENCY_CODE AS ENUM ('USD', 'GEL', 'EUR', 'GBP', 'JPY', 'KRW', 'CNY', 'INR', 'BDT', 'BRL', 'CAD', 'CHF', 'CLP', 'CZK', 'DKK', 'HKD', 'HUF', 'IDR', 'ILS', 'MXN', 'MYR', 'NOK', 'NZD', 'PHP', 'PLN', 'RON', 'SEK', 'SGD', 'THB', 'TRY', 'TWD', 'ZAR');
CREATE TYPE CHARGE_STATUS AS ENUM ('pending', 'succeeded', 'failed');

/* Billings table */
CREATE TABLE billings (
    id BIGSERIAL PRIMARY KEY,
    external_billing_id UUID NOT NULL UNIQUE,
    user_id TEXT NOT NULL,
    description TEXT NOT NULL,
    currency CURRENCY_CODE NOT NULL,
    currency_precision SMALLINT NOT NULL,
    status BILLING_STATUS NOT NULL DEFAULT 'open',
    planned_closed_at TIMESTAMPTZ DEFAULT NULL,
    actual_closed_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX billing_user_id_with_status_idx ON billings (user_id, status);
CREATE INDEX billing_external_billing_id_idx ON billings USING HASH (external_billing_id);

/* Line items table */
CREATE TABLE line_items (
    id BIGSERIAL PRIMARY KEY,
    billing_id BIGSERIAL NOT NULL REFERENCES billings(id),
    description TEXT NOT NULL,
    amount_minor BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

/* Charges table */
CREATE TABLE charges (
    id BIGSERIAL PRIMARY KEY,
    billing_id BIGSERIAL NOT NULL REFERENCES billings(id),
    amount_minor BIGINT NOT NULL,
    status CHARGE_STATUS NOT NULL DEFAULT 'pending',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX charge_billing_id_idx ON charges (billing_id);
