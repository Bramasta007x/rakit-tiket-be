-- gate_logs table
-- Audit trail for gate check-in

DROP TABLE IF EXISTS gate_logs;
DROP TYPE IF EXISTS gate_log_action_enum;

CREATE TYPE gate_log_action_enum AS ENUM (
    'CHECK_IN',
    'CHECK_OUT',
    'INVALID',
    'DUPLICATE',
    'EXCEEDED'
);

CREATE TABLE gate_logs (
    id uuid NOT NULL,
    event_id uuid NOT NULL,

    -- Reference
    physical_ticket_id uuid NOT NULL,
    scanned_by varchar(100) NULL,

    -- Action
    action gate_log_action_enum NOT NULL,
    success bool NOT NULL,
    message varchar(255) NULL,

    -- Location
    gate_name varchar(50) NULL,
    ticket_type varchar(50) NULL,
    scan_sequence int NOT NULL DEFAULT 0,

    -- Metadata
    created_at timestamptz NOT NULL DEFAULT NOW(),

    CONSTRAINT gate_logs_pkey PRIMARY KEY (id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS gate_logs_event_id ON gate_logs(event_id);
CREATE INDEX IF NOT EXISTS gate_logs_physical_ticket_id ON gate_logs(physical_ticket_id);
CREATE INDEX IF NOT EXISTS gate_logs_created_at ON gate_logs(created_at);
CREATE INDEX IF NOT EXISTS gate_logs_gate_name ON gate_logs(gate_name);
CREATE INDEX IF NOT EXISTS gate_logs_action ON gate_logs(action);