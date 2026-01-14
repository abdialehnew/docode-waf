-- Add WebSocket support to vhost_locations table

ALTER TABLE vhost_locations 
ADD COLUMN IF NOT EXISTS websocket_enabled BOOLEAN DEFAULT false;
