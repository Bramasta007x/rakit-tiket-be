CREATE TYPE manual_transfer_status_enum AS ENUM ('PENDING', 'APPROVED', 'REJECTED');

CREATE TABLE manual_transfers (
    id uuid NOT NULL,
    order_id uuid NOT NULL REFERENCES orders(id) ON DELETE CASCADE,
    bank_account_id uuid NOT NULL REFERENCES bank_accounts(id) ON DELETE CASCADE,
    transfer_amount decimal(15,2) NOT NULL,
    transfer_proof_url varchar NOT NULL,
    transfer_proof_filename varchar,
    sender_name varchar NOT NULL,
    sender_account_number varchar,
    transfer_date timestamptz NOT NULL,
    admin_notes text,
    reviewed_by uuid,
    reviewed_at timestamptz,
    status manual_transfer_status_enum NOT NULL DEFAULT 'PENDING',
    deleted bool NOT NULL DEFAULT false,
    data_hash varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_manual_transfers PRIMARY KEY (id)
);

CREATE INDEX idx_manual_transfers_order_id ON manual_transfers (order_id);
CREATE INDEX idx_manual_transfers_status ON manual_transfers (status, deleted);
CREATE INDEX idx_manual_transfers_created_at ON manual_transfers (created_at);
