package models

import (
	"time"
)

// Admin represents an administrator user
type Admin struct {
	ID               string     `json:"id" db:"id"`
	Username         string     `json:"username" db:"username"`
	Email            string     `json:"email" db:"email"`
	PasswordHash     string     `json:"-" db:"password_hash"`
	FullName         string     `json:"full_name,omitempty" db:"full_name"`
	Role             string     `json:"role" db:"role"`
	IsActive         bool       `json:"is_active" db:"is_active"`
	LastLogin        *time.Time `json:"last_login,omitempty" db:"last_login"`
	ResetToken       *string    `json:"-" db:"reset_token"`
	ResetTokenExpiry *time.Time `json:"-" db:"reset_token_expiry"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at" db:"updated_at"`
}

// VHost represents a virtual host configuration
type VHost struct {
	ID               string    `json:"id" db:"id"`
	Name             string    `json:"name" db:"name"`
	Domain           string    `json:"domain" db:"domain"`
	BackendURL       string    `json:"backend_url" db:"backend_url"`
	SSLEnabled       bool      `json:"ssl_enabled" db:"ssl_enabled"`
	SSLCertificateID string    `json:"ssl_certificate_id,omitempty" db:"ssl_certificate_id"`
	SSLCertPath      string    `json:"ssl_cert_path,omitempty" db:"ssl_cert_path"`
	SSLKeyPath       string    `json:"ssl_key_path,omitempty" db:"ssl_key_path"`
	Enabled          bool      `json:"enabled" db:"enabled"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// VHostLocation represents a specific path location within a vhost
type VHostLocation struct {
	ID         string    `json:"id" db:"id"`
	VHostID    string    `json:"vhost_id" db:"vhost_id"`
	Path       string    `json:"path" db:"path"`
	BackendURL string    `json:"backend_url" db:"backend_url"`
	Enabled    bool      `json:"enabled" db:"enabled"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" db:"updated_at"`
}

// IPGroup represents a group of IP addresses
type IPGroup struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description string    `json:"description" db:"description"`
	Type        string    `json:"type" db:"type"` // whitelist, blacklist
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// IPAddress represents an individual IP address in a group
type IPAddress struct {
	ID          string    `json:"id" db:"id"`
	GroupID     string    `json:"group_id" db:"group_id"`
	IPAddress   string    `json:"ip_address" db:"ip_address"`
	CIDRMask    *int      `json:"cidr_mask,omitempty" db:"cidr_mask"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// BlockingRule represents a rule for blocking traffic
type BlockingRule struct {
	ID        string    `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Type      string    `json:"type" db:"type"` // ip, region, url, user_agent
	Pattern   string    `json:"pattern" db:"pattern"`
	Action    string    `json:"action" db:"action"` // block, challenge, allow
	Enabled   bool      `json:"enabled" db:"enabled"`
	Priority  int       `json:"priority" db:"priority"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// RateLimitRule represents a rate limiting rule
type RateLimitRule struct {
	ID                string    `json:"id" db:"id"`
	Name              string    `json:"name" db:"name"`
	PathPattern       string    `json:"path_pattern" db:"path_pattern"`
	RequestsPerSecond int       `json:"requests_per_second" db:"requests_per_second"`
	Burst             int       `json:"burst" db:"burst"`
	Enabled           bool      `json:"enabled" db:"enabled"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
}

// TrafficLog represents a log entry for HTTP traffic
type TrafficLog struct {
	ID           string    `json:"id" db:"id"`
	Timestamp    time.Time `json:"timestamp" db:"timestamp"`
	ClientIP     string    `json:"client_ip" db:"client_ip"`
	Method       string    `json:"method" db:"method"`
	URL          string    `json:"url" db:"url"`
	StatusCode   int       `json:"status_code" db:"status_code"`
	ResponseTime int       `json:"response_time" db:"response_time"`
	BytesSent    int64     `json:"bytes_sent" db:"bytes_sent"`
	UserAgent    string    `json:"user_agent" db:"user_agent"`
	CountryCode  string    `json:"country_code" db:"country_code"`
	Blocked      bool      `json:"blocked" db:"blocked"`
	BlockReason  string    `json:"block_reason,omitempty" db:"block_reason"`
}

// AttackLog represents a log entry for detected attacks
type AttackLog struct {
	ID          string    `json:"id" db:"id"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	ClientIP    string    `json:"client_ip" db:"client_ip"`
	AttackType  string    `json:"attack_type" db:"attack_type"`
	Severity    string    `json:"severity" db:"severity"`
	Description string    `json:"description" db:"description"`
	Blocked     bool      `json:"blocked" db:"blocked"`
	RuleID      *string   `json:"rule_id,omitempty" db:"rule_id"`
}

// SSLCertificate represents an SSL/TLS certificate
type SSLCertificate struct {
	ID        string    `json:"id" db:"id"`
	Domain    string    `json:"domain" db:"domain"`
	CertPath  string    `json:"cert_path" db:"cert_path"`
	KeyPath   string    `json:"key_path" db:"key_path"`
	ExpiresAt time.Time `json:"expires_at" db:"expires_at"`
	AutoRenew bool      `json:"auto_renew" db:"auto_renew"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
