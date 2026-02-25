-- Drop table if exists
DROP TABLE IF EXISTS tickets;

DROP TYPE IF EXISTS ticket_status_enum;

CREATE TYPE ticket_status_enum AS ENUM (
    'AVAILABLE',
    'BOOKOUT',
    'SOLD'
);

CREATE TABLE tickets (
    id uuid NOT NULL,

    -- Ticket Info
    "type" varchar NOT NULL,             -- SILVER, GOLD, FESTIVAL
    title varchar NOT NULL,
    status ticket_status_enum NOT NULL DEFAULT 'AVAILABLE',
    description text NULL,

    -- Pricing & Stock
    price numeric(12, 2) NOT NULL,
    total integer NOT NULL,

    available_qty integer NOT NULL,
    booked_qty integer NOT NULL DEFAULT 0,
    sold_qty integer NOT NULL DEFAULT 0,

    -- Flags & Ordering
    is_presale bool NOT NULL DEFAULT false,
    order_priority integer NOT NULL DEFAULT 0,

    -- Metadata
    deleted bool NOT NULL DEFAULT false,
    data_hash varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NULL,

    CONSTRAINT tickets_pkey PRIMARY KEY (id),

    -- Prevent negative values
    CONSTRAINT tickets_qty_non_negative CHECK (
        total >= 0 AND
        available_qty >= 0 AND
        booked_qty >= 0 AND
        sold_qty >= 0
    ),

    -- Ensure math consistency
    CONSTRAINT tickets_qty_consistency CHECK (
        available_qty + booked_qty + sold_qty = total
    )
);

-- Indexes for Faster Search
CREATE INDEX idx_tickets_type ON tickets ("type");
CREATE INDEX idx_tickets_status ON tickets (status);
CREATE INDEX idx_tickets_is_presale ON tickets (is_presale);
CREATE INDEX idx_tickets_created_at ON tickets (created_at);
