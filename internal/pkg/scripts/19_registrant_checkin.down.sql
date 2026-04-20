-- 19_registrant_checkin.down.sql
-- Remove check-in fields from registrants table

ALTER TABLE public.registrants DROP COLUMN IF EXISTS checked_in;
ALTER TABLE public.registrants DROP COLUMN IF EXISTS checked_in_at;