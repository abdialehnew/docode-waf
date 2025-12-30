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
