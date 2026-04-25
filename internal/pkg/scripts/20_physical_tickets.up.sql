-- physical_tickets table
-- Physical tickets for gate check-in system (standalone - no registrant/attendee relation)

DROP TABLE IF EXISTS physical_tickets;
DROP TYPE IF EXISTS physical_ticket_status_enum;

CREATE TYPE physical_ticket_status_enum AS ENUM (
    'ACTIVE',
    'CHECKED_IN',
    'CHECKED_OUT',
    'EXCEEDED',
    'VOID'
);

CREATE TABLE physical_tickets (
    id uuid NOT NULL,
    event_id uuid NOT NULL REFERENCES events(id) ON DELETE CASCADE,

    -- Ticket reference
    ticket_type varchar(50) NOT NULL,
    ticket_id uuid NOT NULL,

    -- Physical QR Code
    qr_code varchar(50) UNIQUE NOT NULL,
    qr_code_hash varchar(64),

    -- Status
    status physical_ticket_status_enum NOT NULL DEFAULT 'ACTIVE',

    -- Scan Count
    scan_count int NOT NULL DEFAULT 0,

    -- Timestamps
    checked_in_at timestamptz NULL,
    checked_out_at timestamptz NULL,

    -- Metadata
    deleted bool NOT NULL DEFAULT false,
    data_hash varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NULL,

    CONSTRAINT physical_tickets_pkey PRIMARY KEY (id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_physical_tickets_qr_code ON physical_tickets(qr_code);
CREATE INDEX IF NOT EXISTS idx_physical_tickets_event_id ON physical_tickets(event_id);
CREATE INDEX IF NOT EXISTS idx_physical_tickets_status ON physical_tickets(status);
CREATE INDEX IF NOT EXISTS idx_physical_tickets_ticket_type ON physical_tickets(ticket_type);
CREATE INDEX IF NOT EXISTS idx_physical_tickets_ticket_id ON physical_tickets(ticket_id);