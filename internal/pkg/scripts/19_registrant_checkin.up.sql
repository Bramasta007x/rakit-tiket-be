-- 19_registrant_checkin.up.sql
-- Add check-in fields to registrants table

ALTER TABLE public.registrants ADD COLUMN IF NOT EXISTS checked_in BOOLEAN DEFAULT FALSE;
ALTER TABLE public.registrants ADD COLUMN IF NOT EXISTS checked_in_at TIMESTAMP;