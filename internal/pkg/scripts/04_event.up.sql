-- event definition

-- Drop table

-- DROP TABLE event;

DROP TYPE IF EXISTS event_status_enum;

CREATE TYPE event_status_enum AS ENUM (
    'DRAFT',
    'PUBLISHED',
    'COMPLETED',
    'CANCELED'
);

CREATE TABLE events (
    id uuid NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    status event_status_enum NOT NULL DEFAULT 'DRAFT',
    
    ticket_prefix_code VARCHAR(10) NOT NULL,
    max_ticket_per_tx INTEGER NOT NULL DEFAULT 4,
    
    deleted BOOLEAN NOT NULL DEFAULT false,
    data_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT events_pkey PRIMARY KEY (id)
);

-- Indexes for Faster Search
CREATE INDEX idx_events_slug ON events(slug);
CREATE INDEX idx_events_status ON events(status);