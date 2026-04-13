-- Migration: 16_payment_gateway_config.up.sql
-- Description: Add payment_gateways and payment_settings tables for multi-gateway support

-- Payment Gateways table
CREATE TABLE payment_gateways (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    is_enabled BOOLEAN DEFAULT false,
    is_active BOOLEAN DEFAULT false,
    display_order INT DEFAULT 0,
    deleted BOOLEAN DEFAULT FALSE,
    data_hash VARCHAR DEFAULT '-',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Seed default gateways
INSERT INTO payment_gateways (code, name, display_order) VALUES
    ('MIDTRANS', 'Midtrans', 1),
    ('XENDIT', 'Xendit', 2),
    ('DOKU', 'Doku', 3);

-- Payment Settings table
CREATE TABLE payment_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    setting_key VARCHAR(100) UNIQUE NOT NULL,
    setting_value BOOLEAN DEFAULT false,
    display_order INT DEFAULT 0,
    deleted BOOLEAN DEFAULT FALSE,
    data_hash VARCHAR DEFAULT '-',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Manual transfer DISABLED by default
INSERT INTO payment_settings (setting_key, setting_value, display_order) VALUES
    ('MANUAL_TRANSFER_ENABLED', false, 1);
