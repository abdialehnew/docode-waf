-- Create extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Admins table for authentication
CREATE TABLE admins (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    full_name VARCHAR(255),
    role VARCHAR(50) DEFAULT 'admin',
    is_active BOOLEAN DEFAULT true,
    last_login TIMESTAMP,
    reset_token VARCHAR(255),
    reset_token_expiry TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create index on email and username
CREATE INDEX idx_admins_email ON admins(email);
CREATE INDEX idx_admins_username ON admins(username);

-- SSL Certificates table
CREATE TABLE certificates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    cert_content TEXT NOT NULL,
    key_content TEXT NOT NULL,
    common_name VARCHAR(255),
    issuer VARCHAR(255),
    valid_from TIMESTAMP NOT NULL,
    valid_to TIMESTAMP NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_certificates_status ON certificates(status);
CREATE INDEX idx_certificates_valid_to ON certificates(valid_to);

-- Virtual Hosts table
CREATE TABLE vhosts (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    domain VARCHAR(255) NOT NULL,
    backend_url VARCHAR(512) NOT NULL,
    ssl_enabled BOOLEAN DEFAULT false,
    ssl_cert_path VARCHAR(512),
    ssl_key_path VARCHAR(512),
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Virtual Host Locations table
CREATE TABLE vhost_locations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    vhost_id UUID REFERENCES vhosts(id) ON DELETE CASCADE,
    path VARCHAR(512) NOT NULL,
    backend_url VARCHAR(512) NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- IP Groups table
CREATE TABLE ip_groups (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    type VARCHAR(50) NOT NULL, -- 'whitelist', 'blacklist'
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- IP Addresses table
CREATE TABLE ip_addresses (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    group_id UUID REFERENCES ip_groups(id) ON DELETE CASCADE,
    ip_address VARCHAR(45) NOT NULL, -- Support IPv4 and IPv6
    cidr_mask INT, -- For IP blocks (e.g., /24)
    description TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Blocking Rules table
CREATE TABLE blocking_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL, -- 'ip', 'region', 'url', 'user_agent'
    pattern VARCHAR(512) NOT NULL,
    action VARCHAR(50) NOT NULL, -- 'block', 'challenge', 'allow'
    enabled BOOLEAN DEFAULT true,
    priority INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Rate Limit Rules table
CREATE TABLE rate_limit_rules (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(255) NOT NULL,
    path_pattern VARCHAR(512) NOT NULL,
    requests_per_second INT NOT NULL,
    burst INT NOT NULL,
    enabled BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Traffic Logs table
CREATE TABLE traffic_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    client_ip VARCHAR(45) NOT NULL,
    method VARCHAR(10) NOT NULL,
    url TEXT NOT NULL,
    status_code INT NOT NULL,
    response_time INT NOT NULL, -- milliseconds
    bytes_sent BIGINT DEFAULT 0,
    user_agent TEXT,
    country_code VARCHAR(2),
    blocked BOOLEAN DEFAULT false,
    block_reason VARCHAR(255)
);

-- Attack Logs table
CREATE TABLE attack_logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    client_ip VARCHAR(45) NOT NULL,
    attack_type VARCHAR(100) NOT NULL, -- 'http_flood', 'sql_injection', 'xss', 'bot'
    severity VARCHAR(20) NOT NULL, -- 'low', 'medium', 'high', 'critical'
    description TEXT,
    blocked BOOLEAN DEFAULT true,
    rule_id UUID REFERENCES blocking_rules(id) ON DELETE SET NULL
);

-- SSL Certificates table
CREATE TABLE ssl_certificates (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    domain VARCHAR(255) NOT NULL UNIQUE,
    cert_path VARCHAR(512) NOT NULL,
    key_path VARCHAR(512) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    auto_renew BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_traffic_logs_timestamp ON traffic_logs(timestamp DESC);
CREATE INDEX idx_traffic_logs_client_ip ON traffic_logs(client_ip);
CREATE INDEX idx_attack_logs_timestamp ON attack_logs(timestamp DESC);
CREATE INDEX idx_attack_logs_client_ip ON attack_logs(client_ip);
CREATE INDEX idx_ip_addresses_group_id ON ip_addresses(group_id);
CREATE INDEX idx_vhost_locations_vhost_id ON vhost_locations(vhost_id);

-- Create update timestamp trigger
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_vhosts_updated_at BEFORE UPDATE ON vhosts
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_vhost_locations_updated_at BEFORE UPDATE ON vhost_locations
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ip_groups_updated_at BEFORE UPDATE ON ip_groups
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_blocking_rules_updated_at BEFORE UPDATE ON blocking_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_rate_limit_rules_updated_at BEFORE UPDATE ON rate_limit_rules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_ssl_certificates_updated_at BEFORE UPDATE ON ssl_certificates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_certificates_updated_at BEFORE UPDATE ON certificates
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_admins_updated_at BEFORE UPDATE ON admins
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
