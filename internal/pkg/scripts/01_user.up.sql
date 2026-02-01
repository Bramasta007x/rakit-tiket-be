-- "user" definition

-- Drop table
-- DROP TABLE "user";

CREATE TYPE user_role AS ENUM ('ADMIN', 'GROUND STAFF');

CREATE TABLE "user" (
    id uuid NOT NULL,
    name varchar NULL,
    email varchar NULL,
    password_hash varchar NULL,
    role user_role NOT NULL,
    deleted bool NOT NULL DEFAULT false,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NULL,
    CONSTRAINT user_pkey PRIMARY KEY (id),
    CONSTRAINT user_email_un UNIQUE (email)
);

CREATE UNIQUE INDEX user_email_idx ON "user" USING btree (email);
CREATE INDEX user_role_idx ON "user" USING btree (role);
