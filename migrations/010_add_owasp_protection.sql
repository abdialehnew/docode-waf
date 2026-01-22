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
