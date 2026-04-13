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

// stripProtocol removes http://, https:// and trailing slashes from a URL
// for use in nginx upstream server directives
func stripProtocol(url string) string {
	url = strings.TrimPrefix(url, "http://")
	url = strings.TrimPrefix(url, "https://")
	url = strings.TrimSuffix(url, "/")
	return url
}

// parseJSONBackends parses JSON array of backend URLs
func parseJSONBackends(jsonStr string, backends *[]string) error {
	return json.Unmarshal([]byte(jsonStr), backends)
}

type VHostWithLocations struct {
	*models.VHost
	CustomLocations      []CustomLocation
	IPRules              []IPRule
	UpstreamName         string // Sanitized domain name for upstream
	HasUpstream          bool   // True if multiple backends configured
	OCSPStaplingEnabled  bool   // True if certificate has an issuer (not self-signed)
	RateLimitRate        string // Calculated rate (e.g., 10r/s)
	RateLimitBurst       int    // Burst value for limit_req
	GeoIPAvailable       bool   // True if MaxMind database exists
	AppName              string // Application name from settings
}

type IPRule struct {
	Address string
	Action  string // allow or deny
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
{{if .RateLimitEnabled}}
# Rate Limiting Zone for this VHost
limit_req_zone $binary_remote_addr zone=rate_{{sanitizeDomainForUpstream .Domain}}:10m rate={{.RateLimitRate}};
{{end}}
server {
    listen 80;
    server_name {{.Domain}};
    
    # ACME Challenge Support (Let's Encrypt)
    location ^~ /.well-known/acme-challenge/ {
        proxy_pass http://waf:9090;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # Disable caching and security checks for validation
        expires off;
        access_log on;
    }
    
    # Security Headers for HTTP
    {{if .SecurityHeadersEnabled}}add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    {{end}}
    {{if .SSLEnabled}}
    # Redirect HTTP to HTTPS
    return 301 https://$server_name$request_uri;
}

server {
    listen 443 ssl;
    {{if .HTTP3Enabled}}listen 443 quic reuseport;
    http3 on;
    add_header Alt-Svc 'h3=":443"; ma=86400' always;
    {{end}}
    http2 on;
    server_name {{.Domain}};
    {{if .RateLimitEnabled}}
    # Rate Limiting
    limit_req zone=rate_{{sanitizeDomainForUpstream .Domain}} burst={{.RateLimitBurst}} nodelay;
    {{end}}

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
    {{if .OCSPStaplingEnabled}}ssl_stapling on;
    ssl_stapling_verify on;
    resolver 8.8.8.8 8.8.4.4 valid=300s;
    resolver_timeout 5s;
    {{else}}# OCSP Stapling disabled (self-signed or missing issuer)
    # ssl_stapling off;
    {{end}}
    
    # Performance & Security Settings
    {{if .HideServerTokens}}server_tokens off;{{else}}server_tokens on;{{end}}
    {{if .BrotliEnabled}}brotli on;
    brotli_comp_level 6;
    brotli_static on;{{end}}

    # Security Headers
    {{if .HSTSEnabled}}add_header Strict-Transport-Security "max-age={{.HSTSMaxAge}}{{if .HSTSIncludeSubdomains}}; includeSubDomains{{end}}{{if .HSTSPreload}}; preload{{end}}" always;{{end}}
    {{if .SecurityHeadersEnabled}}add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header Referrer-Policy "no-referrer-when-downgrade" always;
    {{if .HSTSEnabled}}# HSTS is enabled via Strict-Transport-Security above{{end}}{{end}}
    add_header Content-Security-Policy "default-src 'self' http: https: data: blob: 'unsafe-inline'; script-src 'self' 'unsafe-inline' 'unsafe-eval' https://challenges.cloudflare.com https://www.google.com https://www.gstatic.com; frame-src 'self' https://challenges.cloudflare.com https://www.google.com; connect-src 'self' https://challenges.cloudflare.com https://www.google.com" always;
    {{end}}

    # Access and Error Logs
    access_log /var/log/nginx/{{.Domain}}_access.log;
    error_log /var/log/nginx/{{.Domain}}_error.log warn;
    
    # IP Access Control
    {{range .IPRules}}{{.Action}} {{.Address}};
    {{end}}
    
    # Custom Error Pages for 403 (IP/Region Block)
    error_page 403 /access-restricted.html;
    location = /access-restricted.html {
        root /usr/share/nginx/html;
        allow all;
        internal;
        sub_filter '{{ "{{" }}IP_ADDRESS{{ "}}" }}' '$remote_addr';
        sub_filter '{{ "{{" }}DOMAIN_NAME{{ "}}" }}' '$host';
        sub_filter 'DoCode WAF' '{{.AppName}}';
        sub_filter 'Docode WAF' '{{.AppName}}';
        sub_filter_once on;
    }
    
    # Region Filtering
    {{if and .RegionFilteringEnabled .GeoIPAvailable}}
    {{if .RegionWhitelist}}
    # Whitelist: Allow only listed countries
    set $allow_country no;
    {{range .RegionWhitelist}}if ($geoip2_data_country_code = "{{.}}") { set $allow_country yes; }
    {{end}}
    if ($allow_country = no) { return 403; }
    {{else if .RegionBlacklist}}
    # Blacklist: Block listed countries
    {{range .RegionBlacklist}}if ($geoip2_data_country_code = "{{.}}") { return 403; }
    {{end}}
    {{end}}
    {{else if .RegionFilteringEnabled}}
    # Region Filtering enabled but GeoIP database is MISSING
    # add_header X-WAF-Warning "GeoIP database not found" always;
    {{end}}
    
    # Per-VHost Upload Size Limit
    client_max_body_size {{.MaxUploadSize}}m;
    client_body_buffer_size {{.ClientBodyBufferSize}}k;
    
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
        {{if .RateLimitEnabled}}limit_req zone=rate_{{sanitizeDomainForUpstream $.Domain}} burst={{$.RateLimitBurst}} nodelay;{{end}}
        
        proxy_pass http://waf:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # HTTP Version & Connection (for keepalive)
        proxy_http_version 1.1;
        proxy_set_header Connection "";
    }
    {{end}}
{{range .CustomLocations}}
    # Custom Location: {{.Path}}
    location {{.Path}} {
        rewrite ^{{.Path}}/?(.*) /$1 break;
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
    # Request Handling based on VHost Type
    location / {
        {{if .RateLimitEnabled}}# Rate Limiting
        limit_req zone=rate_{{sanitizeDomainForUpstream $.Domain}} burst={{$.RateLimitBurst}} nodelay;{{end}}
        
        {{if eq .Type "dead"}}
        # Return 404 for dead hosts
        return 404;
        {{else if eq .Type "redirect"}}
        # Return 301 Redirect
        return 301 {{.BackendURL}}$request_uri;
        {{else}}
        # Proxy Type (Default)
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
        
        # Per-VHost Timeouts
        proxy_connect_timeout {{.ProxyConnectTimeout}}s;
        proxy_send_timeout {{.ProxyConnectTimeout}}s;
        proxy_read_timeout {{.ProxyReadTimeout}}s;
        
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
        {{end}}
    }
}
`

// GenerateVHostConfigContent generates the nginx configuration content string
func (s *NginxConfigService) GenerateVHostConfigContent(data *VHostWithLocations) (string, error) {
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
		"sanitizeDomainForUpstream": sanitizeDomainForUpstream,
	}

	tmpl, err := template.New("vhost").Funcs(funcMap).Parse(VHostTemplate)
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}
	return buf.String(), nil
}

// GenerateVHostConfig generates nginx configuration for a virtual host
func (s *NginxConfigService) GenerateVHostConfig(vhost *models.VHost) error {
	vhostWithLocs := s.prepareVHostWithLocations(vhost)

	if s.db != nil {
		s.fetchCustomLocations(vhost.ID, vhostWithLocs)
		s.fetchIPRules(vhost.ID, vhostWithLocs)

		// Check if OCSP Stapling should be enabled
		if vhost.SSLCertificateID != "" {
			var cert models.Certificate
			err := s.db.Get(&cert, "SELECT issuer, common_name FROM certificates WHERE id = $1", vhost.SSLCertificateID)
			if err == nil {
				// Enable OCSP stapling only if it's not self-signed
				// Basic check: issuer != common_name and issuer is not empty/localhost
				if cert.Issuer != "" && cert.Issuer != cert.CommonName && cert.Issuer != "localhost" {
					vhostWithLocs.OCSPStaplingEnabled = true
				}
			}
		}

		// Fetch App Name for sub_filter
		var appName string
		err := s.db.Get(&appName, "SELECT app_name FROM app_settings WHERE id = 1")
		if err == nil && appName != "" {
			vhostWithLocs.AppName = appName
		} else {
			vhostWithLocs.AppName = "Docode WAF"
		}
	} else {
		vhostWithLocs.AppName = "Docode WAF"
	}

	content, err := s.GenerateVHostConfigContent(vhostWithLocs)
	if err != nil {
		return err
	}

	return s.writeVHostConfigToFile(vhost.Domain, content)
}

func (s *NginxConfigService) prepareVHostWithLocations(vhost *models.VHost) *VHostWithLocations {
	upstreamName := sanitizeDomainForUpstream(vhost.Domain)
	allBackends := []string{}
	for _, b := range vhost.Backends {
		allBackends = append(allBackends, stripProtocol(b))
	}
	if len(allBackends) > 0 && vhost.BackendURL != "" {
		allBackends = append([]string{stripProtocol(vhost.BackendURL)}, allBackends...)
	}

	// Update the actual vhost backends
	vhost.Backends = allBackends

	// Calculate Nginx rate and burst
	rate := "10r/s"
	burst := 100
	if vhost.RateLimitEnabled {
		rate = calculateNginxRate(vhost.RateLimitRequests, vhost.RateLimitWindow)
		burst = vhost.RateLimitRequests
	}

	// Check if GeoIP database exists
	geoIPAvailable := false
	if _, err := os.Stat("/etc/nginx/geoip/GeoLite2-Country.mmdb"); err == nil {
		geoIPAvailable = true
	} else if _, err := os.Stat("./data/nginx/geoip/GeoLite2-Country.mmdb"); err == nil {
		geoIPAvailable = true
	}

	// Set default ClientBodyBufferSize if not provided
	if vhost.ClientBodyBufferSize <= 0 {
		vhost.ClientBodyBufferSize = 128
	}

	return &VHostWithLocations{
		VHost:           vhost,
		CustomLocations: []CustomLocation{},
		IPRules:         []IPRule{},
		UpstreamName:    upstreamName,
		HasUpstream:     len(allBackends) > 0,
		RateLimitRate:   rate,
		RateLimitBurst:  burst,
		GeoIPAvailable:  geoIPAvailable,
	}
}

// calculateNginxRate converts requests and window into nginx rate string (r/s or r/m)
func calculateNginxRate(requests, window int) string {
	if window <= 0 {
		window = 60
	}
	if requests <= 0 {
		requests = 100
	}

	// If window is exactly 1s or less, use r/s
	if window == 1 {
		return fmt.Sprintf("%dr/s", requests)
	}

	// Calculate per-minute rate for better precision in nginx
	// Nginx supports r/s and r/m
	rpm := (requests * 60) / window
	if rpm < 1 {
		rpm = 1 // Minimum 1r/m supported by Nginx
	}

	// If RPM is high enough to be clean per second, use r/s
	if rpm >= 60 && rpm%60 == 0 {
		return fmt.Sprintf("%dr/s", rpm/60)
	}

	return fmt.Sprintf("%dr/m", rpm)
}

func (s *NginxConfigService) fetchCustomLocations(vhostID string, vhostWithLocs *VHostWithLocations) {
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

	if err := s.db.Select(&locations, query, vhostID); err != nil {
		return
	}

	for _, loc := range locations {
		customLoc := CustomLocation{
			Path:              loc.Path,
			ProxyPass:         loc.ProxyPass,
			CustomConfig:      loc.CustomConfig,
			WebSocketEnabled:  loc.WebSocketEnabled,
			LoadBalanceMethod: "round_robin",
		}

		if loc.Backends != nil && *loc.Backends != "" && *loc.Backends != "[]" {
			var backends []string
			if err := parseJSONBackends(*loc.Backends, &backends); err == nil && len(backends) > 0 {
				strippedBackends := make([]string, 0, len(backends))
				for _, b := range backends {
					strippedBackends = append(strippedBackends, stripProtocol(b))
				}
				customLoc.Backends = strippedBackends
				customLoc.HasUpstream = true
				customLoc.UpstreamName = vhostWithLocs.UpstreamName + "_loc_" + sanitizePath(loc.Path)
			}
		}

		if loc.LoadBalanceMethod != nil {
			customLoc.LoadBalanceMethod = *loc.LoadBalanceMethod
		}

		vhostWithLocs.CustomLocations = append(vhostWithLocs.CustomLocations, customLoc)
	}
}

func (s *NginxConfigService) fetchIPRules(vhostID string, vhostWithLocs *VHostWithLocations) {
	var ipRules []struct {
		Address string `db:"ip_address"`
		Mask    *int   `db:"cidr_mask"`
		Type    string `db:"type"`
	}

	ipQuery := `
		SELECT ia.ip_address, ia.cidr_mask, ig.type
		FROM ip_addresses ia
		JOIN ip_groups ig ON ia.group_id = ig.id
		JOIN ip_group_vhosts igv ON ig.id = igv.ip_group_id
		WHERE igv.vhost_id = $1
	`

	if err := s.db.Select(&ipRules, ipQuery, vhostID); err != nil {
		return
	}

	hasWhitelist := false
	for _, rule := range ipRules {
		addr := rule.Address
		if rule.Mask != nil {
			addr = fmt.Sprintf("%s/%d", addr, *rule.Mask)
		}

		action := "deny"
		if rule.Type == "whitelist" {
			action = "allow"
			hasWhitelist = true
		}

		vhostWithLocs.IPRules = append(vhostWithLocs.IPRules, IPRule{
			Address: addr,
			Action:  action,
		})
	}

	if hasWhitelist {
		vhostWithLocs.IPRules = append(vhostWithLocs.IPRules, IPRule{
			Address: "all",
			Action:  "deny",
		})
	}
}

func (s *NginxConfigService) writeVHostConfigToFile(domain, content string) error {
	if err := os.MkdirAll(constants.NginxConfigDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	configPath := filepath.Join(constants.NginxConfigDir, fmt.Sprintf("%s.conf", domain))
	file, err := os.Create(configPath)
	if err != nil {
		return fmt.Errorf("failed to create config file: %w", err)
	}
	defer file.Close()

	if _, err := file.WriteString(content); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
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
