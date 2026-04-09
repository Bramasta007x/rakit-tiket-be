CREATE TYPE payment_method_enum AS ENUM ('GATEWAY', 'MANUAL_TRANSFER');

CREATE TABLE bank_accounts (
    id uuid NOT NULL,
    bank_name varchar NOT NULL,
    bank_code varchar NOT NULL,
    account_number varchar NOT NULL,
    account_holder varchar NOT NULL,
    is_active bool NOT NULL DEFAULT true,
    is_default bool NOT NULL DEFAULT false,
    instruction_text text,
    deleted bool NOT NULL DEFAULT false,
    data_hash varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_bank_accounts PRIMARY KEY (id)
);

CREATE INDEX idx_bank_accounts_active ON bank_accounts (is_active, deleted);
