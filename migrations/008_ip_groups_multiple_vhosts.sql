-- Migration 008: Enable multiple vhosts per IP group
-- Creates a junction table for many-to-many relationship between ip_groups and vhosts

-- Create junction table
CREATE TABLE IF NOT EXISTS ip_group_vhosts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_group_id UUID NOT NULL REFERENCES ip_groups(id) ON DELETE CASCADE,
    vhost_id UUID NOT NULL REFERENCES vhosts(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(ip_group_id, vhost_id)
);

-- Create index for faster lookups
CREATE INDEX IF NOT EXISTS idx_ip_group_vhosts_group_id ON ip_group_vhosts(ip_group_id);
CREATE INDEX IF NOT EXISTS idx_ip_group_vhosts_vhost_id ON ip_group_vhosts(vhost_id);

-- Migrate existing data from ip_groups.vhost_id to junction table
INSERT INTO ip_group_vhosts (ip_group_id, vhost_id)
SELECT id, vhost_id 
FROM ip_groups 
WHERE vhost_id IS NOT NULL
ON CONFLICT (ip_group_id, vhost_id) DO NOTHING;

-- Keep vhost_id column for backward compatibility (will be deprecated)
-- In future migrations, we can safely remove it after verifying all systems use the junction table
-- ALTER TABLE ip_groups DROP COLUMN IF EXISTS vhost_id;
