-- Add vhost_id to ip_groups table to support per-vhost blacklist/whitelist
ALTER TABLE ip_groups 
ADD COLUMN vhost_id UUID REFERENCES vhosts(id) ON DELETE CASCADE;

-- Add index for faster lookups
CREATE INDEX idx_ip_groups_vhost_id ON ip_groups(vhost_id);

-- Allow NULL for global rules (backward compatibility)
COMMENT ON COLUMN ip_groups.vhost_id IS 'NULL = global rule, otherwise specific vhost only';
