-- Add 'rejected' status to registrants table
-- Required for manual transfer rejection flow

ALTER TABLE registrants DROP CONSTRAINT IF EXISTS registrants_status_check;
ALTER TABLE registrants ADD CONSTRAINT registrants_status_check 
    CHECK (status IN ('pending', 'paid', 'failed', 'cancelled', 'rejected'));
