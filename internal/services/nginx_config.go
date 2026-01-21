package services

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/aleh/docode-waf/internal/constants"
	"github.com/aleh/docode-waf/internal/models"
	"github.com/jmoiron/sqlx"
)

type NginxConfigService struct {
	db *sqlx.DB
}

func NewNginxConfigService() *NginxConfigService {
	return &NginxConfigService{}
}

func NewNginxConfigServiceWithDB(db *sqlx.DB) *NginxConfigService {
	return &NginxConfigService{db: db}
}

// sanitizeDomainForUpstream converts domain name to valid nginx upstream name
// e.g., "example.com" -> "example_com"
func sanitizeDomainForUpstream(domain string) string {
	// Replace dots and hyphens with underscores
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return re.ReplaceAllString(domain, "_")
}

// sanitizePath converts path to valid upstream suffix
// e.g., "/api/v1" -> "api_v1"
func sanitizePath(path string) string {
	// Remove leading slash and replace special chars
	path = strings.TrimPrefix(path, "/")
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	return re.ReplaceAllString(path, "_")
}

// parseJSONBackends parses JSON array of backend URLs
func parseJSONBackends(jsonStr string, backends *[]string) error {
	return json.Unmarshal([]byte(jsonStr), backends)
}

type VHostWithLocations struct {
	*models.VHost
	CustomLocations []CustomLocation
	UpstreamName    string // Sanitized domain name for upstream
	HasUpstream     bool   // True if multiple backends configured
}

type CustomLocation struct {
	Path              string
	ProxyPass         string
	CustomConfig      string
	WebSocketEnabled  bool
	Backends          []string
	LoadBalanceMethod string
	UpstreamName      string // Unique upstream name for this location
	HasUpstream       bool
}

// VHostTemplate is the nginx configuration template for a virtual host
const VHostTemplate = `# Virtual Host: {{.Name}}
# Generated automatically - Optimized for Performance & Security
{{if .HasUpstream}}
# Upstream for load balancing: {{.Domain}}
upstream {{.UpstreamName}}_backend {
    {{if eq .LoadBalanceMethod "least_conn"}}least_conn;
    {{else if eq .LoadBalanceMethod "ip_hash"}}ip_hash;
    {{end}}{{range .Backends}}
    server {{.}} weight=1 max_fails=3 fail_timeout=30s;{{end}}
    
    # Keepalive connections to backend
    keepalive 32;
    keepalive_requests 1000;
    keepalive_timeout 60s;
}
{{end}}{{range .CustomLocations}}{{if .HasUpstream}}
# Upstream for location: {{.Path}}
upstream {{.UpstreamName}}_backend {
    {{if eq .LoadBalanceMethod "least_conn"}}least_conn;
    {{else if eq .LoadBalanceMethod "ip_hash"}}ip_hash;
    {{end}}{{range .Backends}}
    server {{.}} weight=1 max_fails=3 fail_timeout=30s;{{end}}
    
    keepalive 16;
    keepalive_requests 500;
    keepalive_timeout 60s;
}
{{end}}{{end}}
server {
    listen 80;
    server_name {{.Domain}};
    
    # Security Headers for HTTP
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;

    {{if .SSLEnabled}}
    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl;
    http2 on;
    server_name {{.Domain}};

    # SSL Configuration
    ssl_certificate /etc/nginx/ssl/certificates/{{.SSLCertificateID}}/cert.pem;
    ssl_certificate_key /etc/nginx/ssl/certificates/{{.SSLCertificateID}}/key.pem;
    
    # SSL Security - Modern Configuration
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers 'ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:DHE-RSA-AES128-GCM-SHA256:DHE-RSA-AES256-GCM-SHA384';
    ssl_prefer_server_ciphers off;
    
    # SSL Session Optimization
    ssl_session_cache shared:SSL:50m;
    ssl_session_timeout 1d;
    ssl_session_tickets off;
    
    # OCSP Stapling
    ssl_stapling on;
    ssl_stapling_verify on;
    resolver 8.8.8.8 8.8.4.4 valid=300s;
    resolver_timeout 5s;
    
    # Security Headers
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://challenges.cloudflare.com https://www.google.com https://www.gstatic.com; frame-src 'self' https://challenges.cloudflare.com https://www.google.com; connect-src 'self' https://challenges.cloudflare.com https://www.google.com" always;
    {{end}}

    # Access and Error Logs
    access_log /var/log/nginx/{{.Domain}}_access.log;
    error_log /var/log/nginx/{{.Domain}}_error.log warn;
    
    # Per-VHost Upload Size Limit
    client_max_body_size 100m;
    
    # Static Assets Caching (Performance Optimization)
    location ~* \.(jpg|jpeg|png|gif|ico|svg|webp|avif)$ {
        expires 30d;
        add_header Cache-Control "public, no-transform, immutable";
        add_header Vary "Accept-Encoding";
        access_log off;
        log_not_found off;
        
        proxy_pass http://waf:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_cache backend_cache;
        proxy_cache_valid 200 30d;
        proxy_cache_valid 404 1m;
    }
    
    location ~* \.(css|js|woff|woff2|ttf|eot)$ {
        expires 7d;
        add_header Cache-Control "public, no-transform";
        add_header Vary "Accept-Encoding";
        access_log off;
        
        proxy_pass http://waf:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_cache backend_cache;
        proxy_cache_valid 200 7d;
    }
    
    # Security: Deny access to hidden files (except .well-known and .vite for dev)
    location ~ /\.(?!well-known|vite) {
        deny all;
        access_log off;
        log_not_found off;
    }
    
    # Security: Deny access to sensitive files
    location ~* \.(git|svn|htaccess|htpasswd|env)$ {
        deny all;
        access_log off;
        log_not_found off;
    }
    {{if not (hasAPILocation .CustomLocations)}}
    # Rate Limiting for API endpoints
    location ~* ^/api/ {
        limit_req zone=api burst=200 nodelay;
        limit_req_status 429;
        
        proxy_pass http://waf:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Proxy Buffering
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
        
        # Timeouts
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # HTTP Version & Connection
        proxy_http_version 1.1;
        proxy_set_header Connection "";
    }
    {{end}}
{{range .CustomLocations}}
    # Custom Location: {{.Path}}
    location {{.Path}} {
        {{if .HasUpstream}}proxy_pass http://{{.UpstreamName}}_backend;
        {{else if .ProxyPass}}proxy_pass {{.ProxyPass}};
        {{end}}{{if or .HasUpstream .ProxyPass}}proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        {{if .WebSocketEnabled}}
        # WebSocket Support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        {{else}}proxy_set_header Connection "";
        {{end}}{{end}}{{if .CustomConfig}}
        {{.CustomConfig}}{{end}}
    }
{{end}}
    # Proxy to WAF - All requests go through WAF middleware first
    location / {
        # Rate limiting
        limit_req zone=general burst=100 nodelay;
        
        {{if .HasUpstream}}proxy_pass http://{{.UpstreamName}}_backend;
        {{else}}proxy_pass http://waf:8080;
        {{end}}proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Forwarded-Host $host;
        proxy_set_header X-Forwarded-Port $server_port;
        
        # WebSocket Support
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Per-VHost Timeouts (longer for API/heavy responses)
        proxy_connect_timeout 120s;
        proxy_send_timeout 120s;
        proxy_read_timeout 120s;
        
        # Proxy Cache Configuration
        proxy_cache backend_cache;
        proxy_cache_valid 200 10m;
        proxy_cache_valid 404 1m;
        proxy_cache_bypass $http_upgrade $http_cache_control;
        proxy_no_cache $http_pragma $http_authorization;
        add_header X-Cache-Status $upstream_cache_status always;
        {{if .CustomConfig}}
        # Custom Configuration
        {{.CustomConfig}}{{end}}
    }
}
`

// GenerateVHostConfig generates nginx configuration for a virtual host
func (s *NginxConfigService) GenerateVHostConfig(vhost *models.VHost) error {
	// Create template with helper function
	funcMap := template.FuncMap{
		"hasAPILocation": func(locations []CustomLocation) bool {
			for _, loc := range locations {
				if loc.Path == "/api/" || loc.Path == "/api" || loc.Path == "~* ^/api/" {
					return true
				}
			}
			return false
		},
	}

	tmpl, err := template.New("vhost").Funcs(funcMap).Parse(VHostTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse template: %w", err)
	}

	// Prepare vhost with locations
	upstreamName := sanitizeDomainForUpstream(vhost.Domain)
	vhostWithLocs := &VHostWithLocations{
		VHost:           vhost,
		CustomLocations: []CustomLocation{},
		UpstreamName:    upstreamName,
		HasUpstream:     len(vhost.Backends) > 0,
	}

	// Fetch custom locations from database if db is available
	if s.db != nil {
		var locations []struct {
			ID                string  `db:"id"`
			Path              string  `db:"path"`
			ProxyPass         string  `db:"proxy_pass"`
			CustomConfig      string  `db:"custom_config"`
			WebSocketEnabled  bool    `db:"websocket_enabled"`
			Backends          *string `db:"backends"`
			LoadBalanceMethod *string `db:"load_balance_method"`
		}

		query := `
			SELECT id, path, COALESCE(proxy_pass, '') as proxy_pass, COALESCE(custom_config, '') as custom_config, 
				   COALESCE(websocket_enabled, false) as websocket_enabled,
				   backends::text as backends, COALESCE(load_balance_method, 'round_robin') as load_balance_method
			FROM vhost_locations
			WHERE vhost_id = $1 AND enabled = true
			ORDER BY length(path) DESC
		`

		if err := s.db.Select(&locations, query, vhost.ID); err == nil {
			for _, loc := range locations {
				customLoc := CustomLocation{
					Path:              loc.Path,
					ProxyPass:         loc.ProxyPass,
					CustomConfig:      loc.CustomConfig,
					WebSocketEnabled:  loc.WebSocketEnabled,
					LoadBalanceMethod: "round_robin",
				}

				// Parse backends from JSON if present
				if loc.Backends != nil && *loc.Backends != "" && *loc.Backends != "[]" {
					var backends []string
					if err := parseJSONBackends(*loc.Backends, &backends); err == nil && len(backends) > 0 {
						customLoc.Backends = backends
						customLoc.HasUpstream = true
						customLoc.UpstreamName = upstreamName + "_loc_" + sanitizePath(loc.Path)
					}
				}

				if loc.LoadBalanceMethod != nil {
					customLoc.LoadBalanceMethod = *loc.LoadBalanceMethod
				}

				vhostWithLocs.CustomLocations = append(vhostWithLocs.CustomLocations, customLoc)
			}
		}
	}

	// Create config directory if not exists
	if err := os.MkdirAll(constants.NginxConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Create config file
	configPath := filepath.Join(constants.NginxConfigDir, fmt.Sprintf("%s.conf", vhost.Domain))
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, vhostWithLocs); err != nil {
		return fmt.Errorf("failed to execute template: %w", err)
	}

	return nil
}

// DeleteVHostConfig deletes nginx configuration for a virtual host
func (s *NginxConfigService) DeleteVHostConfig(domain string) error {
	configPath := filepath.Join(constants.NginxConfigDir, fmt.Sprintf("%s.conf", domain))
	if err := os.Remove(configPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete config file: %w", err)
	}
	return nil
}

// RegenerateAllVHostConfigs regenerates all nginx configurations
func (s *NginxConfigService) RegenerateAllVHostConfigs(vhosts []*models.VHost) error {
	// Clean up existing configs
	files, err := filepath.Glob(filepath.Join(constants.NginxConfigDir, "*.conf"))
	if err != nil {
		return fmt.Errorf("failed to list config files: %w", err)
	}

	for _, file := range files {
		if err := os.Remove(file); err != nil {
			return fmt.Errorf("failed to remove config file %s: %w", file, err)
		}
	}

	// Generate new configs
	for _, vhost := range vhosts {
		if err := s.GenerateVHostConfig(vhost); err != nil {
			return fmt.Errorf("failed to generate config for %s: %w", vhost.Domain, err)
		}
	}

	return nil
}
