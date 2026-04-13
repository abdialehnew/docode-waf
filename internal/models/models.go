package models

import (
	"encoding/json"
	"time"

	"github.com/lib/pq"
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
	ID                     string    `json:"id" db:"id"`
	Name                   string    `json:"name" db:"name"`
	Type                   string    `json:"type" db:"type"` // proxy, redirect, dead, stream
	Domain                 string    `json:"domain" db:"domain"`
	BackendURL             string    `json:"backend_url" db:"backend_url"`
	Backends               []string  `json:"backends" db:"-"` // Multiple backend URLs for load balancing
	LoadBalanceMethod      string    `json:"load_balance_method" db:"load_balance_method"`
	CustomConfig           string    `json:"custom_config" db:"custom_config"`
	SSLEnabled             bool      `json:"ssl_enabled" db:"ssl_enabled"`
	SSLCertificateID       string    `json:"ssl_certificate_id,omitempty" db:"ssl_certificate_id"`
	SSLCertPath            string    `json:"ssl_cert_path,omitempty" db:"ssl_cert_path"`
	SSLKeyPath             string    `json:"ssl_key_path,omitempty" db:"ssl_key_path"`
	Enabled                bool      `json:"enabled" db:"enabled"`
	RegionWhitelist        pq.StringArray `json:"region_whitelist" db:"region_whitelist"`
	RegionBlacklist        pq.StringArray `json:"region_blacklist" db:"region_blacklist"`
	RegionFilteringEnabled bool           `json:"region_filtering_enabled" db:"region_filtering_enabled"`
	DefenseMode            string         `json:"defense_mode" db:"defense_mode"`
	MaxUploadSize          int            `json:"max_upload_size" db:"max_upload_size"`
	ProxyReadTimeout       int            `json:"proxy_read_timeout" db:"proxy_read_timeout"`
	ProxyConnectTimeout    int            `json:"proxy_connect_timeout" db:"proxy_connect_timeout"`
	BotDetectionEnabled    bool           `json:"bot_detection_enabled" db:"bot_detection_enabled"`
	BotDetectionType       string         `json:"bot_detection_type" db:"bot_detection_type"`
	RecaptchaVersion       string         `json:"recaptcha_version" db:"recaptcha_version"`
	RateLimitEnabled       bool           `json:"rate_limit_enabled" db:"rate_limit_enabled"`
	RateLimitRequests      int            `json:"rate_limit_requests" db:"rate_limit_requests"`
	RateLimitWindow        int            `json:"rate_limit_window" db:"rate_limit_window"`
	CacheEnabled           bool           `json:"cache_enabled" db:"cache_enabled"`
	CacheTTL               int            `json:"cache_ttl" db:"cache_ttl"`
	CacheMethods           pq.StringArray `json:"cache_methods" db:"cache_methods"`
	CacheIgnoreHeaders     bool           `json:"cache_ignore_headers" db:"cache_ignore_headers"`
	HSTSEnabled            bool            `json:"hsts_enabled" db:"hsts_enabled"`
	HSTSMaxAge             int             `json:"hsts_max_age" db:"hsts_max_age"`
	HSTSIncludeSubdomains  bool            `json:"hsts_include_subdomains" db:"hsts_include_subdomains"`
	HSTSPreload            bool            `json:"hsts_preload" db:"hsts_preload"`
	BrotliEnabled          bool            `json:"brotli_enabled" db:"brotli_enabled"`
	HTTP3Enabled           bool            `json:"http3_enabled" db:"http3_enabled"`
	HideServerTokens       bool            `json:"hide_server_tokens" db:"hide_server_tokens"`
	SecurityHeadersEnabled bool            `json:"security_headers_enabled" db:"security_headers_enabled"`
	ClientBodyBufferSize   int             `json:"client_body_buffer_size" db:"client_body_buffer_size"`
	WebsocketEnabled       bool            `json:"websocket_enabled" db:"websocket_enabled"`
	HTTPVersion            string          `json:"http_version" db:"http_version"`
	TLSVersion             string          `json:"tls_version" db:"tls_version"`
	CustomHeaders          json.RawMessage `json:"custom_headers" db:"custom_headers"`
	CreatedAt              time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt              time.Time       `json:"updated_at" db:"updated_at"`
}

// VHostLocation represents a specific path location within a vhost
type VHostLocation struct {
	ID                string   `json:"id" db:"id"`
	VHostID           string   `json:"vhost_id" db:"vhost_id"`
	Path              string   `json:"path" db:"path"`
	BackendURL        string   `json:"backend_url" db:"backend_url"`
	Backends          []string `json:"backends" db:"-"` // Multiple backend URLs for load balancing
	LoadBalanceMethod string   `json:"load_balance_method" db:"load_balance_method"`
	Enabled           bool     `json:"enabled" db:"enabled"`
	CreatedAt         time.Time `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time `json:"updated_at" db:"updated_at"`
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
