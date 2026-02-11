-- Drop table if exists
DROP TABLE IF EXISTS tickets;

CREATE TABLE tickets (
    id uuid NOT NULL,

    -- Ticket Info
    "type" varchar NOT NULL,             -- SILVER, GOLD, FESTIVAL
    title varchar NOT NULL,
    description text NULL,

    -- Pricing & Stock
    price numeric(12, 2) NOT NULL,
    total integer NOT NULL,
    remaining integer NOT NULL,

    -- Flags & Ordering
    is_presale bool NOT NULL DEFAULT false,
    order_priority integer NOT NULL DEFAULT 0,

    -- Metadata
    deleted bool NOT NULL DEFAULT false,
    data_hash varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NULL,

    CONSTRAINT tickets_pkey PRIMARY KEY (id)
);

-- Indexes for Faster Search
CREATE INDEX idx_tickets_type ON tickets ("type");
CREATE INDEX idx_tickets_is_presale ON tickets (is_presale);
CREATE INDEX idx_tickets_created_at ON tickets (created_at);
