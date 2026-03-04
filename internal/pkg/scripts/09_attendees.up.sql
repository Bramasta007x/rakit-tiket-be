-- attendees definition

-- Drop table

-- DROP TABLE attendees;
CREATE TABLE attendees (
    id uuid NOT NULL,

    -- Relation
    event_id UUID NOT NULL REFERENCES events(id) ON DELETE CASCADE,
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

    CONSTRAINT attendees_pkey PRIMARY KEY (id)
);

-- Indexes for Faster Search
CREATE INDEX idx_attendees_event_id ON attendees(event_id);
CREATE INDEX idx_attendees_registrant_id ON attendees(registrant_id);
CREATE INDEX idx_attendees_ticket_id ON attendees(ticket_id);
CREATE INDEX idx_attendees_deleted ON attendees(deleted);