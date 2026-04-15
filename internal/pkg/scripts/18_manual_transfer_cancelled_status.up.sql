-- Add 'CANCELLED' status to manual_transfer_status_enum

ALTER TYPE manual_transfer_status_enum ADD VALUE IF NOT EXISTS 'CANCELLED';
