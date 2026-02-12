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
