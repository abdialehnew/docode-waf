-- Migration: Add defense_mode to vhosts
-- Description: Adds a defense_mode column to configure the WAF behavior for each vhost

ALTER TABLE vhosts ADD COLUMN IF NOT EXISTS defense_mode VARCHAR(20) DEFAULT 'defense';
