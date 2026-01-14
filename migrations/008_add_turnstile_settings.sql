-- Migration: Add Turnstile settings to app_settings table
-- This allows enabling/disabling Turnstile on login and authentication pages

-- Add turnstile_enabled column to control whether Turnstile is shown (default: disabled)
ALTER TABLE app_settings 
ADD COLUMN IF NOT EXISTS turnstile_enabled BOOLEAN DEFAULT false;

-- Add turnstile_login_enabled column to control Turnstile on login page specifically (default: disabled)
ALTER TABLE app_settings 
ADD COLUMN IF NOT EXISTS turnstile_login_enabled BOOLEAN DEFAULT false;

-- Add turnstile_register_enabled column to control Turnstile on register page specifically (default: disabled)
ALTER TABLE app_settings 
ADD COLUMN IF NOT EXISTS turnstile_register_enabled BOOLEAN DEFAULT false;

-- Update existing row with default values (disabled by default)
UPDATE app_settings 
SET turnstile_enabled = false,
    turnstile_login_enabled = false,
    turnstile_register_enabled = false
WHERE id = 1 
  AND turnstile_enabled IS NULL;
