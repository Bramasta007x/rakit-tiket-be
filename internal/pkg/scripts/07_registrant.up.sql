-- Tabel Registrants
CREATE TABLE registrants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    unique_code VARCHAR(255) NOT NULL UNIQUE,
    ticket_id UUID REFERENCES tickets(id), -- Optional jika order bisa multi-ticket type, tapi sesuaikan kebutuhan
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
);