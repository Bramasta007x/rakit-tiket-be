-- Remove 'rejected' status from registrants table

ALTER TABLE registrants DROP CONSTRAINT IF EXISTS registrants_status_check;
ALTER TABLE registrants ADD CONSTRAINT registrants_status_check 
    CHECK (status IN ('pending', 'paid', 'failed', 'cancelled'));
