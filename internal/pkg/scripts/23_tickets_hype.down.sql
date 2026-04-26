-- Rollback tickets_hype extension

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

DROP INDEX IF EXISTS idx_tickets_sale_start;
DROP INDEX IF EXISTS idx_tickets_sale_end;
DROP INDEX IF EXISTS idx_tickets_flash;
DROP INDEX IF EXISTS idx_tickets_flash_time;