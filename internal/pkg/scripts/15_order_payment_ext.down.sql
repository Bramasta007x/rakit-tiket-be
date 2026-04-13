ALTER TABLE orders DROP COLUMN IF EXISTS payment_type;
ALTER TABLE orders DROP COLUMN IF EXISTS payment_proof_url;
ALTER TABLE orders DROP COLUMN IF EXISTS payment_proof_filename;
ALTER TABLE orders DROP COLUMN IF EXISTS verified_by;
ALTER TABLE orders DROP COLUMN IF EXISTS verified_at;
