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
