-- Drop table
-- DROP TABLE landing_page_configs;

CREATE TABLE landing_page_configs (
    id uuid NOT NULL,
    banner_image varchar NULL,
    venue_image varchar NULL,
    event_creator varchar NULL,
    event_name varchar NOT NULL,
    event_date varchar NOT NULL,
    event_time_start varchar NOT NULL,
    event_time_end varchar NOT NULL,
    event_location varchar NULL,
    terms_and_conditions jsonb NOT NULL DEFAULT '[]'::jsonb,

    deleted bool NOT NULL DEFAULT false,
    data_hash varchar NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NULL,

    CONSTRAINT landing_page_configs_pkey PRIMARY KEY (id)
);
