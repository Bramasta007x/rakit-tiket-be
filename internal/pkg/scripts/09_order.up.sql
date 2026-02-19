CREATE TABLE orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    registrant_id UUID NOT NULL REFERENCES registrants(id) ON DELETE CASCADE,
    order_number VARCHAR(255) NOT NULL UNIQUE,
    amount NUMERIC(12, 2) NOT NULL DEFAULT 0,
    currency VARCHAR(10) DEFAULT 'IDR',
    
    -- Generic Payment Gateway Info
    payment_gateway VARCHAR(50), -- Contoh: 'MIDTRANS', 'XENDIT', 'STRIPE', 'MANUAL'
    payment_method VARCHAR(50), -- Contoh: 'bank_transfer', 'ewallet', 'credit_card'
    payment_channel VARCHAR(50), -- Contoh: 'bca_va', 'gopay'
    payment_status VARCHAR(20) DEFAULT 'pending', -- pending, paid, failed, expired
    
    -- Gateway Data
    payment_token VARCHAR(255), -- Universal token (Snap Token Midtrans / Invoice ID Xendit)
    payment_url VARCHAR(500), -- Universal URL redirect ke halaman pembayaran
    payment_transaction_id VARCHAR(255), -- Transaction ID internal dari gateway
    payment_metadata JSON, -- Data spesifik yang berbeda tiap gateway bisa dilempar ke sini
    
    payment_time TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    
    -- Metadata
    deleted BOOLEAN DEFAULT FALSE,
    data_hash VARCHAR DEFAULT '-',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);