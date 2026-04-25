-- gate_configs table
-- Configuration for gate check-in mode

DROP TABLE IF EXISTS gate_configs;
DROP TYPE IF EXISTS gate_mode_enum;

CREATE TYPE gate_mode_enum AS ENUM (
    'CHECK_IN',
    'CHECK_IN_OUT'
);

CREATE TABLE gate_configs (
    id uuid NOT NULL,
    event_id uuid UNIQUE NOT NULL REFERENCES events(id) ON DELETE CASCADE,

    -- Mode Configuration
    mode gate_mode_enum NOT NULL DEFAULT 'CHECK_IN',

    -- Max Scan per Ticket (only for CHECK_IN mode)
    max_scan_per_ticket int NOT NULL DEFAULT 1,

    -- Override per Category (JSON)
    max_scan_by_type jsonb NULL,

    -- Active
    is_active bool NOT NULL DEFAULT true,

    -- Metadata
    deleted bool NOT NULL DEFAULT false,
    data_hash varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NULL,

    CONSTRAINT gate_configs_pkey PRIMARY KEY (id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS gate_configs_event_id ON gate_configs(event_id);