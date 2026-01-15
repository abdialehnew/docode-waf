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
