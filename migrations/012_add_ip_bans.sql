CREATE TABLE IF NOT EXISTS ip_bans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ip_address VARCHAR(45) NOT NULL,
    reason VARCHAR(255) NOT NULL,
    banned_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    violation_count INT DEFAULT 1,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_ip_bans_ip_address ON ip_bans(ip_address);
CREATE INDEX idx_ip_bans_expires_at ON ip_bans(expires_at);

-- Trigger to update updated_at
CREATE TRIGGER update_ip_bans_updated_at
    BEFORE UPDATE ON ip_bans
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
