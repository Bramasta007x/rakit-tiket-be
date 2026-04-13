ALTER TABLE orders ADD COLUMN payment_type varchar DEFAULT NULL;

ALTER TABLE orders ADD COLUMN payment_proof_url varchar DEFAULT NULL;
ALTER TABLE orders ADD COLUMN payment_proof_filename varchar DEFAULT NULL;

ALTER TABLE orders ADD COLUMN verified_by uuid DEFAULT NULL;
ALTER TABLE orders ADD COLUMN verified_at timestamptz DEFAULT NULL;
