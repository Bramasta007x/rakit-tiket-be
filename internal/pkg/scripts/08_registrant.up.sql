-- registrants definition

-- Drop table

-- DROP TABLE registrants;
CREATE TABLE registrants (
    id uuid NOT NULL,

    -- Relation
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
    ticket_id UUID REFERENCES tickets(id),

    unique_code VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) NOT NULL,
    phone VARCHAR(50) NOT NULL,
    gender VARCHAR(10),
    birthdate DATE,
    total_cost NUMERIC(12, 2) DEFAULT 0,
    total_tickets INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'pending' CHECK (status IN ('pending', 'paid', 'failed', 'cancelled')),
   
    -- Metadata
    deleted bool NOT NULL DEFAULT false,
    data_hash varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NULL,

    CONSTRAINT registrants_pkey PRIMARY KEY (id)
);

-- Indexes for Faster Search
CREATE INDEX IF NOT EXISTS idx_registrants_event_id ON registrants(event_id);
CREATE INDEX IF NOT EXISTS idx_registrants_ticket_id ON registrants(ticket_id);
CREATE INDEX IF NOT EXISTS idx_registrants_email ON registrants(email);
CREATE INDEX IF NOT EXISTS idx_registrants_status ON registrants(status);
CREATE INDEX IF NOT EXISTS idx_registrants_created_at ON registrants(created_at);
CREATE INDEX IF NOT EXISTS idx_registrants_deleted ON registrants(deleted);