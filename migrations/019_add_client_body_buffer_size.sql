-- Migration: Add client_body_buffer_size to vhosts table
-- Description: Adds a column for configuring Nginx client_body_buffer_size per vhost
-- Default: 1024 (1MB)

ALTER TABLE vhosts 
ADD COLUMN IF NOT EXISTS client_body_buffer_size INT DEFAULT 1024;

COMMENT ON COLUMN vhosts.client_body_buffer_size IS 'Nginx client_body_buffer_size in KB';
