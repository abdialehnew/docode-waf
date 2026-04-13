-- Source: init.sql
-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Admins table for authentication
CREATE TABLE admins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'admin',
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    reset_token VARCHAR(255),
    reset_token_expiry TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email and username
CREATE INDEX idx_admins_email ON admins(email);
CREATE INDEX idx_admins_username ON admins(username);

-- SSL Certificates table
CREATE TABLE certificates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    cert_content TEXT NOT NULL,
    key_content TEXT NOT NULL,
    common_name VARCHAR(255),
    issuer VARCHAR(255),
    valid_from TIMESTAMP NOT NULL,
    valid_to TIMESTAMP NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_certificates_status ON certificates(status);
CREATE INDEX idx_certificates_valid_to ON certificates(valid_to);

-- Virtual Hosts table
CREATE TABLE vhosts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    domain VARCHAR(255) NOT NULL,
    backend_url VARCHAR(512) NOT NULL,
    ssl_enabled BOOLEAN DEFAULT false,
    ssl_cert_path VARCHAR(512),
    ssl_key_path VARCHAR(512),
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Virtual Host Locations table
CREATE TABLE vhost_locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vhost_id UUID REFERENCES vhosts(id) ON DELETE CASCADE,
    path VARCHAR(512) NOT NULL,
    backend_url VARCHAR(512) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- IP Groups table
CREATE TABLE ip_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- 'whitelist', 'blacklist'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- IP Addresses table
CREATE TABLE ip_addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    group_id UUID REFERENCES ip_groups(id) ON DELETE CASCADE,
    ip_address VARCHAR(45) NOT NULL, -- Support IPv4 and IPv6
    cidr_mask INT, -- For IP blocks (e.g., /24)
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Blocking Rules table
CREATE TABLE blocking_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'ip', 'region', 'url', 'user_agent'
    pattern VARCHAR(512) NOT NULL,
    action VARCHAR(50) NOT NULL, -- 'block', 'challenge', 'allow'
    enabled BOOLEAN DEFAULT true,
    priority INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Rate Limit Rules table
CREATE TABLE rate_limit_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    path_pattern VARCHAR(512) NOT NULL,
    requests_per_second INT NOT NULL,
    burst INT NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Traffic Logs table
CREATE TABLE traffic_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    client_ip VARCHAR(45) NOT NULL,
    method VARCHAR(10) NOT NULL,
    url TEXT NOT NULL,
    status_code INT NOT NULL,
    response_time INT NOT NULL, -- milliseconds
    bytes_sent BIGINT DEFAULT 0,
    user_agent TEXT,
    country_code VARCHAR(2),
    blocked BOOLEAN DEFAULT false,
    block_reason VARCHAR(255)
);

-- Attack Logs table
CREATE TABLE attack_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    client_ip VARCHAR(45) NOT NULL,
    attack_type VARCHAR(100) NOT NULL, -- 'http_flood', 'sql_injection', 'xss', 'bot'
    severity VARCHAR(20) NOT NULL, -- 'low', 'medium', 'high', 'critical'
    description TEXT,
    blocked BOOLEAN DEFAULT true,
    rule_id UUID REFERENCES blocking_rules(id) ON DELETE SET NULL
);

-- SSL Certificates table
CREATE TABLE ssl_certificates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    domain VARCHAR(255) NOT NULL UNIQUE,
    cert_path VARCHAR(512) NOT NULL,
    key_path VARCHAR(512) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    auto_renew BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_traffic_logs_timestamp ON traffic_logs(timestamp DESC);
CREATE INDEX idx_traffic_logs_client_ip ON traffic_logs(client_ip);
CREATE INDEX idx_attack_logs_timestamp ON attack_logs(timestamp DESC);
CREATE INDEX idx_attack_logs_client_ip ON attack_logs(client_ip);
CREATE INDEX idx_ip_addresses_group_id ON ip_addresses(group_id);
CREATE INDEX idx_vhost_locations_vhost_id ON vhost_locations(vhost_id);

-- Create update timestamp trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_vhosts_updated_at BEFORE UPDATE ON vhosts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vhost_locations_updated_at BEFORE UPDATE ON vhost_locations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ip_groups_updated_at BEFORE UPDATE ON ip_groups
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_blocking_rules_updated_at BEFORE UPDATE ON blocking_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rate_limit_rules_updated_at BEFORE UPDATE ON rate_limit_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ssl_certificates_updated_at BEFORE UPDATE ON ssl_certificates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_certificates_updated_at BEFORE UPDATE ON certificates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_admins_updated_at BEFORE UPDATE ON admins
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- Source: 002_add_vhost_advanced_fields.sql
-- Add advanced fields to vhosts table

-- Add WebSocket support
ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS websocket_enabled BOOLEAN DEFAULT false;

-- Add HTTP version support
ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS http_version VARCHAR(20) DEFAULT 'http/1.1';

-- Add TLS version support for SSL/TLS protocol
ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS tls_version VARCHAR(50) DEFAULT 'TLSv1.2';

-- Add max upload size (in MB)
ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS max_upload_size INT DEFAULT 10;

-- Add proxy timeout settings
ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS proxy_read_timeout INT DEFAULT 60;

ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS proxy_connect_timeout INT DEFAULT 60;

-- Add custom headers (stored as JSON)
ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS custom_headers JSONB DEFAULT '{}'::jsonb;

-- Add SSL certificate reference
ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS ssl_certificate_id UUID REFERENCES certificates(id) ON DELETE SET NULL;

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_vhosts_domain ON vhosts(domain);
CREATE INDEX IF NOT EXISTS idx_vhosts_enabled ON vhosts(enabled);
CREATE INDEX IF NOT EXISTS idx_vhosts_ssl_certificate_id ON vhosts(ssl_certificate_id);

-- Update vhost_locations to support custom nginx config
ALTER TABLE vhost_locations 
ADD COLUMN IF NOT EXISTS proxy_pass VARCHAR(512);

ALTER TABLE vhost_locations 
ADD COLUMN IF NOT EXISTS custom_config TEXT;

-- Create index on vhost_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_vhost_locations_vhost_id ON vhost_locations(vhost_id);


-- Source: 002_add_vhost_to_ipgroups.sql
-- Add vhost_id to ip_groups table to support per-vhost blacklist/whitelist
ALTER TABLE ip_groups 
ADD COLUMN vhost_id UUID REFERENCES vhosts(id) ON DELETE CASCADE;

-- Add index for faster lookups
CREATE INDEX idx_ip_groups_vhost_id ON ip_groups(vhost_id);

-- Allow NULL for global rules (backward compatibility)
COMMENT ON COLUMN ip_groups.vhost_id IS 'NULL = global rule, otherwise specific vhost only';


-- Source: 003_add_app_settings.sql
-- Migration: Add app_settings table
-- Created: 2024-01-15

-- Create app_settings table for application configuration
CREATE TABLE IF NOT EXISTS app_settings (
    id INT PRIMARY KEY DEFAULT 1,
    app_name VARCHAR(255) NOT NULL DEFAULT 'Docode WAF',
    app_logo TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT single_row CHECK (id = 1)
);

-- Insert default settings
INSERT INTO app_settings (id, app_name, app_logo, created_at, updated_at)
VALUES (1, 'Docode WAF', NULL, NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Create index
CREATE INDEX IF NOT EXISTS idx_app_settings_id ON app_settings(id);

-- Add trigger for updated_at
CREATE TRIGGER update_app_settings_updated_at BEFORE UPDATE ON app_settings
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- Source: 003_add_vhost_features.sql
-- Add bot detection and rate limiter settings to vhosts table
ALTER TABLE vhosts 
ADD COLUMN bot_detection_enabled BOOLEAN DEFAULT false,
ADD COLUMN bot_detection_type VARCHAR(50) DEFAULT 'turnstile', -- turnstile, captcha, slide_puzzle
ADD COLUMN rate_limit_enabled BOOLEAN DEFAULT false,
ADD COLUMN rate_limit_requests INT DEFAULT 100,
ADD COLUMN rate_limit_window INT DEFAULT 60; -- in seconds

-- Add indexes for performance
CREATE INDEX idx_vhosts_bot_detection ON vhosts(bot_detection_enabled);
CREATE INDEX idx_vhosts_rate_limit ON vhosts(rate_limit_enabled);

COMMENT ON COLUMN vhosts.bot_detection_enabled IS 'Enable bot detection for this vhost';
COMMENT ON COLUMN vhosts.bot_detection_type IS 'Type of bot detection: turnstile, captcha, slide_puzzle';
COMMENT ON COLUMN vhosts.rate_limit_enabled IS 'Enable rate limiting for this vhost';
COMMENT ON COLUMN vhosts.rate_limit_requests IS 'Maximum requests allowed per window';
COMMENT ON COLUMN vhosts.rate_limit_window IS 'Time window in seconds';


-- Source: 004_add_attack_detection_to_traffic_logs.sql
-- Migration: Add attack detection fields to traffic_logs
-- This allows tracking attacks directly in traffic logs instead of separate attack_logs table

-- Add attack detection flag
ALTER TABLE traffic_logs 
ADD COLUMN IF NOT EXISTS is_attack BOOLEAN DEFAULT false;

-- Add attack type for categorization
ALTER TABLE traffic_logs 
ADD COLUMN IF NOT EXISTS attack_type VARCHAR(100);

-- Add host/domain field to identify which vhost received the request
ALTER TABLE traffic_logs 
ADD COLUMN IF NOT EXISTS host VARCHAR(255);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_traffic_logs_is_attack ON traffic_logs(is_attack);
CREATE INDEX IF NOT EXISTS idx_traffic_logs_attack_type ON traffic_logs(attack_type);
CREATE INDEX IF NOT EXISTS idx_traffic_logs_host ON traffic_logs(host);
CREATE INDEX IF NOT EXISTS idx_traffic_logs_country_code ON traffic_logs(country_code);

-- Add composite index for dashboard queries (timestamp + is_attack)
CREATE INDEX IF NOT EXISTS idx_traffic_logs_timestamp_attack ON traffic_logs(timestamp DESC, is_attack);


-- Source: 005_seed_default_admin.sql
-- Migration: Seed default admin account
-- Created: 2025-12-31
-- Default credentials:
--   Username: admin
--   Password: Admin123!
--   Email: admin@docode.local

-- Insert default admin user
-- Password hash for "Admin123!" generated with bcrypt cost 10
INSERT INTO admins (
    id,
    username,
    email,
    password_hash,
    full_name,
    role,
    is_active,
    created_at,
    updated_at
)
VALUES (
    '00000000-0000-0000-0000-000000000001'::uuid,
    'admin',
    'admin@docode.local',
    '$2a$10$ivX.3Bp.cZaROB6ZTe901eC3FTb/3EpCSrD1XqYnbmydlhoW3DSO.',
    'System Administrator',
    'superadmin',
    true,
    NOW(),
    NOW()
)
ON CONFLICT (username) DO NOTHING;

-- Note: ON CONFLICT ensures this is idempotent and safe to run multiple times



-- Source: 006_add_recaptcha_version.sql
-- Migration: Add recaptcha_version column to vhosts table
-- This allows admin to choose between reCAPTCHA v2 (checkbox) or v3 (score-based)

-- Add recaptcha_version column (default to v2 for backward compatibility)
ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS recaptcha_version VARCHAR(10) DEFAULT 'v2';

-- Add check constraint to ensure only v2 or v3 values
ALTER TABLE vhosts 
ADD CONSTRAINT check_recaptcha_version 
CHECK (recaptcha_version IN ('v2', 'v3'));

-- Create index for faster filtering
CREATE INDEX IF NOT EXISTS idx_vhosts_recaptcha_version 
ON vhosts(recaptcha_version);

-- Update existing rows to have v2 as default
UPDATE vhosts 
SET recaptcha_version = 'v2' 
WHERE recaptcha_version IS NULL;

-- Add comment
COMMENT ON COLUMN vhosts.recaptcha_version IS 'reCAPTCHA version: v2 (checkbox) or v3 (score-based, invisible)';


-- Source: 006_add_smtp_and_signup_settings.sql
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


-- Source: 007_add_region_filtering.sql
-- Migration: Add region-based filtering per vhost
-- Description: Adds columns for whitelisting and blacklisting countries/regions per vhost
-- Author: System
-- Date: 2026-01-05

-- Add region filtering columns to vhosts table
ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS region_whitelist TEXT[] DEFAULT '{}',
ADD COLUMN IF NOT EXISTS region_blacklist TEXT[] DEFAULT '{}',
ADD COLUMN IF NOT EXISTS region_filtering_enabled BOOLEAN DEFAULT FALSE;

-- Add comment for documentation
COMMENT ON COLUMN vhosts.region_whitelist IS 'Array of ISO 3166-1 alpha-2 country codes to whitelist (e.g., ["US", "GB", "ID"]). If not empty, only these countries are allowed.';
COMMENT ON COLUMN vhosts.region_blacklist IS 'Array of ISO 3166-1 alpha-2 country codes to blacklist (e.g., ["CN", "RU"]). These countries will be blocked.';
COMMENT ON COLUMN vhosts.region_filtering_enabled IS 'Enable or disable region-based filtering for this vhost';

-- Note: Region filtering logic:
-- 1. If region_filtering_enabled = false, skip all checks
-- 2. If whitelist is not empty, ONLY allow countries in whitelist
-- 3. If whitelist is empty and blacklist is not empty, block countries in blacklist
-- 4. If both are empty, allow all countries


-- Source: 007_add_websocket_to_vhost_locations.sql
-- Add WebSocket support to vhost_locations table

ALTER TABLE vhost_locations 
ADD COLUMN IF NOT EXISTS websocket_enabled BOOLEAN DEFAULT false;


-- Source: 008_add_turnstile_settings.sql
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


-- Source: 008_ip_groups_multiple_vhosts.sql
-- Migration 008: Enable multiple vhosts per IP group
-- Creates a junction table for many-to-many relationship between ip_groups and vhosts

-- Create junction table
CREATE TABLE IF NOT EXISTS ip_group_vhosts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_group_id UUID NOT NULL REFERENCES ip_groups(id) ON DELETE CASCADE,
    vhost_id UUID NOT NULL REFERENCES vhosts(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ip_group_id, vhost_id)
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_ip_group_vhosts_group_id ON ip_group_vhosts(ip_group_id);
CREATE INDEX IF NOT EXISTS idx_ip_group_vhosts_vhost_id ON ip_group_vhosts(vhost_id);

-- Migrate existing data from ip_groups.vhost_id to junction table
INSERT INTO ip_group_vhosts (ip_group_id, vhost_id)
SELECT id, vhost_id 
FROM ip_groups 
WHERE vhost_id IS NOT NULL
ON CONFLICT (ip_group_id, vhost_id) DO NOTHING;

-- Keep vhost_id column for backward compatibility (will be deprecated)
-- In future migrations, we can safely remove it after verifying all systems use the junction table
-- ALTER TABLE ip_groups DROP COLUMN IF EXISTS vhost_id;


-- Source: 009_add_multiple_backends.sql
-- Migration: Add multiple backends and custom config support
-- This enables load balancing with multiple backend servers

-- Add backends JSONB column to vhosts for multiple backend URLs
ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS backends JSONB DEFAULT '[]';

-- Add load_balance_method column to vhosts
-- Options: round_robin (default), least_conn, ip_hash
ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS load_balance_method VARCHAR(50) DEFAULT 'round_robin';

-- Add custom_config column to vhosts for custom nginx config
ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS custom_config TEXT;

-- Add backends JSONB column to vhost_locations for multiple backend URLs
ALTER TABLE vhost_locations 
ADD COLUMN IF NOT EXISTS backends JSONB DEFAULT '[]';

-- Add load_balance_method column to vhost_locations
ALTER TABLE vhost_locations 
ADD COLUMN IF NOT EXISTS load_balance_method VARCHAR(50) DEFAULT 'round_robin';


-- Source: 010_add_owasp_protection.sql
-- Migration: Add OWASP Top 10 protection fields to vhosts table
-- Date: 2026-01-22

-- OWASP Protection settings
ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS owasp_protection_enabled BOOLEAN DEFAULT true;
ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS owasp_protection_level VARCHAR(20) DEFAULT 'medium';
-- Levels: 'low' (block critical only), 'medium' (block critical+high), 'high' (block critical+high+medium), 'paranoid' (block all)

-- Brute Force Protection settings
ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS brute_force_enabled BOOLEAN DEFAULT false;
ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS brute_force_threshold INTEGER DEFAULT 5;
ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS brute_force_window INTEGER DEFAULT 300;
ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS brute_force_lockout INTEGER DEFAULT 900;
ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS login_paths TEXT DEFAULT '/login,/auth,/signin,/api/login,/api/auth';

-- Security Headers settings
ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS security_headers_enabled BOOLEAN DEFAULT true;
ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS hsts_enabled BOOLEAN DEFAULT true;
ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS hsts_max_age INTEGER DEFAULT 31536000;
ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS csp_policy TEXT DEFAULT '';
ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS permissions_policy TEXT DEFAULT '';

-- Add attack severity to traffic_logs
ALTER TABLE traffic_logs ADD COLUMN IF NOT EXISTS attack_severity VARCHAR(20) DEFAULT '';

-- Add index for attack queries
CREATE INDEX IF NOT EXISTS idx_traffic_logs_attack_severity ON traffic_logs(attack_severity) WHERE attack_severity != '';
CREATE INDEX IF NOT EXISTS idx_traffic_logs_attack_type ON traffic_logs(attack_type) WHERE attack_type != '';


-- Source: 011_add_vhost_type.sql
-- Add type column to vhosts table
ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS type VARCHAR(20) DEFAULT 'proxy';

-- Update existing records if needed (optional, default handles it)
-- UPDATE vhosts SET type = 'proxy' WHERE type IS NULL;


-- Source: 012_add_ip_bans.sql
CREATE TABLE IF NOT EXISTS ip_bans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_address VARCHAR(45) NOT NULL,
    reason VARCHAR(255) NOT NULL,
    banned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    violation_count INT DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ip_bans_ip_address ON ip_bans(ip_address);
CREATE INDEX idx_ip_bans_expires_at ON ip_bans(expires_at);

-- Trigger to update updated_at
CREATE TRIGGER update_ip_bans_updated_at
    BEFORE UPDATE ON ip_bans
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();


-- Source: 013_add_notification_channels.sql
-- Migration: Add notification channels
-- Created: 2026-02-04

CREATE TABLE IF NOT EXISTS notification_channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'email', 'slack', 'discord', 'webhook'
    config JSONB NOT NULL DEFAULT '{}', -- Stores webhook_url, email_address, etc.
    events JSONB NOT NULL DEFAULT '[]', -- Array of event names
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookups
CREATE INDEX IF NOT EXISTS idx_notification_channels_enabled ON notification_channels(enabled);

-- Trigger for updated_at
CREATE TRIGGER update_notification_channels_updated_at BEFORE UPDATE ON notification_channels
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- Source: 014_add_rules_table.sql
-- Migration: Add rules table for visual rule builder
-- Created: 2026-02-05

CREATE TABLE IF NOT EXISTS rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    priority INTEGER NOT NULL DEFAULT 0,
    action VARCHAR(50) NOT NULL, -- 'block', 'allow', 'challenge', 'log'
    conditions JSONB NOT NULL DEFAULT '[]', -- Array of conditions
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookup by priority and enabled status
CREATE INDEX IF NOT EXISTS idx_rules_priority_enabled ON rules(enabled, priority DESC);

-- Trigger for updated_at
CREATE TRIGGER update_rules_updated_at BEFORE UPDATE ON rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();


-- Source: 015_add_match_logic.sql
-- Migration: Add match_logic to rules table
-- Created: 2026-02-05

ALTER TABLE rules ADD COLUMN IF NOT EXISTS match_logic VARCHAR(10) DEFAULT 'AND';


-- Source: 016_add_defense_mode.sql
-- Migration: Add defense_mode to vhosts
-- Description: Adds a defense_mode column to configure the WAF behavior for each vhost

ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS defense_mode VARCHAR(20) DEFAULT 'defense';


-- Source: 017_add_vhost_caching.sql
-- Migration: Add caching support to vhosts
-- Description: Adds columns for configuring caching behavior per vhost

ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS cache_enabled BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS cache_ttl INTEGER DEFAULT 60,
ADD COLUMN IF NOT EXISTS cache_methods TEXT[] DEFAULT '{"GET"}',
ADD COLUMN IF NOT EXISTS cache_ignore_headers BOOLEAN DEFAULT false;

COMMENT ON COLUMN vhosts.cache_enabled IS 'Enable or disable caching for this vhost';
COMMENT ON COLUMN vhosts.cache_ttl IS 'Cache time-to-live in seconds';
COMMENT ON COLUMN vhosts.cache_methods IS 'HTTP methods to cache (e.g., ["GET", "HEAD"])';
COMMENT ON COLUMN vhosts.cache_ignore_headers IS 'Ignore Vary, Cache-Control, and Expires headers from backend';


-- Source: 018_add_advanced_security_performance.sql
-- Migration: Add advanced security and performance fields to vhosts
-- Description: Adds columns for HSTS, Brotli, HTTP/3, and Server Tokens

ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS hsts_include_subdomains BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS hsts_preload BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS brotli_enabled BOOLEAN DEFAULT true,
ADD COLUMN IF NOT EXISTS http3_enabled BOOLEAN DEFAULT false,
ADD COLUMN IF NOT EXISTS hide_server_tokens BOOLEAN DEFAULT true;


-- Source: 019_add_client_body_buffer_size.sql
-- Migration: Add client_body_buffer_size to vhosts table
ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS client_body_buffer_size INT DEFAULT 128;
COMMENT ON COLUMN vhosts.client_body_buffer_size IS 'Nginx client_body_buffer_size in KB';

