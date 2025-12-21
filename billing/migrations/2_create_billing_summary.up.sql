CREATE TABLE billing_summaries (
    id BIGSERIAL PRIMARY KEY,
    external_billing_id UUID NOT NULL UNIQUE,
    summary JSONB NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc', now()),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT timezone('utc', now())
);
