-- "file" definition

-- Drop table

-- DROP TABLE "file";


CREATE TABLE file (
    id uuid NOT NULL,
    relation_id uuid NOT NULL,     
    relation_source varchar NOT NULL,   
    file_name varchar NOT NULL,         
    file_path varchar NOT NULL,                
    file_mime varchar NOT NULL,   
    "description" TEXT NULL,

    deleted bool NOT NULL DEFAULT false,
    data_hash varchar(128) NOT NULL,            
    created_at timestamptz NULL,
    updated_at timestamptz NULL,
    CONSTRAINT file_pkey PRIMARY KEY(id),
    CONSTRAINT file_unique UNIQUE (relation_id, relation_source)
);

CREATE UNIQUE INDEX dile_data_hash_idx ON file USING btree (data_hash);
CREATE INDEX file_relation_id_idx ON file USING btree (relation_id);
CREATE INDEX file_relation_source_idx ON file USING btree (relation_source);
CREATE INDEX file_name_idx ON file USING btree (file_name);
CREATE INDEX file_mime_source_idx ON file USING btree (file_mime);