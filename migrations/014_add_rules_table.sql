-- Migration: Add rules table for visual rule builder
-- Created: 2026-02-05

CREATE TABLE IF NOT EXISTS rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    priority INTEGER NOT NULL DEFAULT 0,
    action VARCHAR(50) NOT NULL, -- 'block', 'allow', 'challenge', 'log'
    conditions JSONB NOT NULL DEFAULT '[]', -- Array of conditions
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Index for faster lookup by priority and enabled status
CREATE INDEX IF NOT EXISTS idx_rules_priority_enabled ON rules(enabled, priority DESC);

-- Trigger for updated_at
CREATE TRIGGER update_rules_updated_at BEFORE UPDATE ON rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
