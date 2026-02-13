-- Tabel Attendees
CREATE TABLE attendees (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    registrant_id UUID NOT NULL REFERENCES registrants(id) ON DELETE CASCADE,
    ticket_id UUID NOT NULL REFERENCES tickets(id),
    name VARCHAR(255) NOT NULL,
    gender VARCHAR(10),
    birthdate DATE,
    
    -- Metadata
    deleted bool NOT NULL DEFAULT false,
    data_hash varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NULL,
);