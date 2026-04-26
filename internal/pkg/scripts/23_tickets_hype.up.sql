-- tickets_hype extension
-- Add urgency, FOMO, and flash sale features to tickets

-- Drop existing columns if any (for idempotency)
ALTER TABLE tickets DROP COLUMN IF EXISTS sale_start_time;
ALTER TABLE tickets DROP COLUMN IF EXISTS sale_end_time;
ALTER TABLE tickets DROP COLUMN IF EXISTS is_flash_sale;
ALTER TABLE tickets DROP COLUMN IF EXISTS flash_sale_price;
ALTER TABLE tickets DROP COLUMN IF EXISTS flash_start_time;
ALTER TABLE tickets DROP COLUMN IF EXISTS flash_end_time;
ALTER TABLE tickets DROP COLUMN IF EXISTS low_stock_threshold;
ALTER TABLE tickets DROP COLUMN IF EXISTS show_stock_alert;
ALTER TABLE tickets DROP COLUMN IF EXISTS show_countdown;
ALTER TABLE tickets DROP COLUMN IF EXISTS countdown_end;

-- Add columns
ALTER TABLE tickets ADD COLUMN sale_start_time timestamptz NULL;
ALTER TABLE tickets ADD COLUMN sale_end_time timestamptz NULL;

-- Flash Sale
ALTER TABLE tickets ADD COLUMN is_flash_sale bool NOT NULL DEFAULT false;
ALTER TABLE tickets ADD COLUMN flash_sale_price numeric(12, 2) NULL;
ALTER TABLE tickets ADD COLUMN flash_start_time timestamptz NULL;
ALTER TABLE tickets ADD COLUMN flash_end_time timestamptz NULL;

-- Stock Alert (20% threshold default)
ALTER TABLE tickets ADD COLUMN low_stock_threshold int NULL;
ALTER TABLE tickets ADD COLUMN show_stock_alert bool NOT NULL DEFAULT true;

-- FOMO / Urgency
ALTER TABLE tickets ADD COLUMN show_countdown bool NOT NULL DEFAULT true;
ALTER TABLE tickets ADD COLUMN countdown_end timestamptz NULL;

-- Indexes
CREATE INDEX IF NOT EXISTS idx_tickets_sale_start ON tickets(sale_start_time);
CREATE INDEX IF NOT EXISTS idx_tickets_sale_end ON tickets(sale_end_time);
CREATE INDEX IF NOT EXISTS idx_tickets_flash ON tickets(is_flash_sale);
CREATE INDEX IF NOT EXISTS idx_tickets_flash_time ON tickets(flash_start_time, flash_end_time);