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
