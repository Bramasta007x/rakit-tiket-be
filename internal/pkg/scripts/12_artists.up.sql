CREATE TABLE artists (
    id uuid NOT NULL,
    image VARCHAR(500),
    name VARCHAR(255) NOT NULL,
    genre VARCHAR(100) NOT NULL,
    artist_social_media jsonb NOT NULL DEFAULT '[]'::jsonb,
    
    deleted BOOLEAN NOT NULL DEFAULT false,
    data_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT artists_pkey PRIMARY KEY (id)
);

CREATE INDEX idx_artists_name ON artists(name);
CREATE INDEX idx_artists_genre ON artists(genre);
CREATE INDEX idx_artists_deleted ON artists(deleted);
