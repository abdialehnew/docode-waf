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
