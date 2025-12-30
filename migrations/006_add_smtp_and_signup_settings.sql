-- Migration: Add signup and SMTP settings
-- Created: 2025-12-31

-- Add signup enable/disable field
ALTER TABLE app_settings 
ADD COLUMN IF NOT EXISTS signup_enabled BOOLEAN DEFAULT true;

-- Add SMTP configuration fields
ALTER TABLE app_settings 
ADD COLUMN IF NOT EXISTS smtp_host VARCHAR(255);

ALTER TABLE app_settings 
ADD COLUMN IF NOT EXISTS smtp_port INT DEFAULT 587;

ALTER TABLE app_settings 
ADD COLUMN IF NOT EXISTS smtp_username VARCHAR(255);

ALTER TABLE app_settings 
ADD COLUMN IF NOT EXISTS smtp_password VARCHAR(255);

ALTER TABLE app_settings 
ADD COLUMN IF NOT EXISTS smtp_from_email VARCHAR(255);

ALTER TABLE app_settings 
ADD COLUMN IF NOT EXISTS smtp_from_name VARCHAR(255) DEFAULT 'Docode WAF';

ALTER TABLE app_settings 
ADD COLUMN IF NOT EXISTS smtp_use_tls BOOLEAN DEFAULT true;

-- Update existing row with defaults
UPDATE app_settings 
SET 
    signup_enabled = true,
    smtp_port = 587,
    smtp_from_name = 'Docode WAF',
    smtp_use_tls = true
WHERE id = 1;
