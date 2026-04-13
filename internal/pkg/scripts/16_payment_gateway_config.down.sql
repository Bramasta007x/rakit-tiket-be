-- Migration: 16_payment_gateway_config.down.sql
-- Description: Remove payment_gateways and payment_settings tables

DROP TABLE IF EXISTS payment_settings;
DROP TABLE IF EXISTS payment_gateways;
