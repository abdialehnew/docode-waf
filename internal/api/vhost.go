package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/aleh/docode-waf/internal/constants"
	"github.com/aleh/docode-waf/internal/models"
	"github.com/aleh/docode-waf/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const (
	proxyReloadWarningMsg = "Warning: Failed to reload proxy map: %v\n"
)

// VHostHandler handles virtual host requests
type ProxyReloader interface {
	ReloadVHosts() error
}

type VHostHandler struct {
	db                 *sqlx.DB
	nginxConfigService *services.NginxConfigService
	vhostService       *services.VHostService
	certService        *services.CertificateService
	proxyReloader      ProxyReloader
}

// NewVHostHandler creates a new vhost handler
func NewVHostHandler(db *sqlx.DB, nginxConfigService *services.NginxConfigService, vhostService *services.VHostService, certService *services.CertificateService, proxyReloader ProxyReloader) *VHostHandler {
	return &VHostHandler{
		db:                 db,
		nginxConfigService: nginxConfigService,
		vhostService:       vhostService,
		certService:        certService,
		proxyReloader:      proxyReloader,
	}
}

// ListVHosts returns all virtual hosts
func (h *VHostHandler) ListVHosts(c *gin.Context) {
	type VHost struct {
		ID                  string          `db:"id" json:"id"`
		Name                string          `db:"name" json:"name"`
		Type                string          `db:"type" json:"type"`
		Domain              string          `db:"domain" json:"domain"`
		BackendURL          string          `db:"backend_url" json:"backend_url"`
		Backends            *string         `db:"backends" json:"backends"`
		LoadBalanceMethod   *string         `db:"load_balance_method" json:"load_balance_method"`
		CustomConfig        *string         `db:"custom_config" json:"custom_config"`
		SSLEnabled          bool            `db:"ssl_enabled" json:"ssl_enabled"`
		SSLCertificateID    *string         `db:"ssl_certificate_id" json:"ssl_certificate_id"`
		SSLCertPath         *string         `db:"ssl_cert_path" json:"ssl_cert_path"`
		SSLKeyPath          *string         `db:"ssl_key_path" json:"ssl_key_path"`
		Enabled             bool            `db:"enabled" json:"enabled"`
		WebsocketEnabled    bool            `db:"websocket_enabled" json:"websocket_enabled"`
		HTTPVersion         string          `db:"http_version" json:"http_version"`
		TLSVersion          string          `db:"tls_version" json:"tls_version"`
		MaxUploadSize       int             `db:"max_upload_size" json:"max_upload_size"`
		ProxyReadTimeout    int             `db:"proxy_read_timeout" json:"proxy_read_timeout"`
		ProxyConnectTimeout int             `db:"proxy_connect_timeout" json:"proxy_connect_timeout"`
		BotDetectionEnabled bool            `db:"bot_detection_enabled" json:"bot_detection_enabled"`
		BotDetectionType    string          `db:"bot_detection_type" json:"bot_detection_type"`
		RecaptchaVersion    string          `db:"recaptcha_version" json:"recaptcha_version"`
		RateLimitEnabled    bool            `db:"rate_limit_enabled" json:"rate_limit_enabled"`
		RateLimitRequests   int             `db:"rate_limit_requests" json:"rate_limit_requests"`
		RateLimitWindow     int             `db:"rate_limit_window" json:"rate_limit_window"`
		DefenseMode         string          `db:"defense_mode" json:"defense_mode"`
		CustomHeaders       json.RawMessage `db:"custom_headers" json:"custom_headers"`
		CreatedAt           time.Time       `db:"created_at" json:"created_at"`
		UpdatedAt           time.Time       `db:"updated_at" json:"updated_at"`
	}

	var vhosts []VHost

	query := `
		SELECT id::text, name, COALESCE(type, 'proxy') as type, domain, backend_url, 
		       backends::text as backends, COALESCE(load_balance_method, 'round_robin') as load_balance_method, custom_config,
		       ssl_enabled, ssl_certificate_id::text, ssl_cert_path, ssl_key_path, enabled,
		       websocket_enabled, http_version, tls_version, max_upload_size,
		       proxy_read_timeout, proxy_connect_timeout,
		       bot_detection_enabled, bot_detection_type, recaptcha_version,
		       rate_limit_enabled, rate_limit_requests, rate_limit_window,
		       COALESCE(defense_mode, 'defense') as defense_mode, custom_headers, created_at, updated_at
		FROM vhosts 
		ORDER BY created_at DESC
	`

	err := h.db.Select(&vhosts, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Query custom locations for all vhosts
	type CustomLocation struct {
		VHostID           string  `db:"vhost_id"`
		Path              string  `db:"path" json:"path"`
		ProxyPass         *string `db:"proxy_pass" json:"proxy_pass"`
		CustomConfig      *string `db:"custom_config" json:"config"`
		WebSocketEnabled  bool    `db:"websocket_enabled" json:"websocket_enabled"`
		Backends          *string `db:"backends" json:"backends"`
		LoadBalanceMethod *string `db:"load_balance_method" json:"load_balance_method"`
	}
	var allLocations []CustomLocation
	locQuery := `
		SELECT vhost_id::text, path, proxy_pass, custom_config, COALESCE(websocket_enabled, false) as websocket_enabled,
		       backends::text as backends, COALESCE(load_balance_method, 'round_robin') as load_balance_method
		FROM vhost_locations
		WHERE enabled = true
		ORDER BY created_at ASC
	`
	_ = h.db.Select(&allLocations, locQuery)

	// Map locations to vhosts
	locationMap := make(map[string][]map[string]interface{})
	for _, loc := range allLocations {
		if _, exists := locationMap[loc.VHostID]; !exists {
			locationMap[loc.VHostID] = []map[string]interface{}{}
		}
		// Parse location backends
		var locBackends []string
		if loc.Backends != nil && *loc.Backends != "" && *loc.Backends != "[]" {
			_ = json.Unmarshal([]byte(*loc.Backends), &locBackends)
		}
		if locBackends == nil {
			locBackends = []string{}
		}

		locationMap[loc.VHostID] = append(locationMap[loc.VHostID], map[string]interface{}{
			"path":                loc.Path,
			"proxy_pass":          loc.ProxyPass,
			"config":              loc.CustomConfig,
			"websocket_enabled":   loc.WebSocketEnabled,
			"backends":            locBackends,
			"load_balance_method": loc.LoadBalanceMethod,
		})
	}

	// Build response with custom_locations
	var response []map[string]interface{}
	for _, vhost := range vhosts {
		customLocs := locationMap[vhost.ID]
		if customLocs == nil {
			customLocs = []map[string]interface{}{}
		}

		backends := parseJSONStringArray(vhost.Backends)
		loadBalanceMethod := constants.DefaultLoadBalanceMethod
		if vhost.LoadBalanceMethod != nil {
			loadBalanceMethod = *vhost.LoadBalanceMethod
		}

		customConfig := ""
		if vhost.CustomConfig != nil {
			customConfig = *vhost.CustomConfig
		}

		response = append(response, map[string]interface{}{
			"id":                    vhost.ID,
			"name":                  vhost.Name,
			"type":                  vhost.Type,
			"domain":                vhost.Domain,
			"backend_url":           vhost.BackendURL,
			"backends":              backends,
			"load_balance_method":   loadBalanceMethod,
			"custom_config":         customConfig,
			"ssl_enabled":           vhost.SSLEnabled,
			"ssl_certificate_id":    vhost.SSLCertificateID,
			"ssl_cert_path":         vhost.SSLCertPath,
			"ssl_key_path":          vhost.SSLKeyPath,
			"enabled":               vhost.Enabled,
			"websocket_enabled":     vhost.WebsocketEnabled,
			"http_version":          vhost.HTTPVersion,
			"tls_version":           vhost.TLSVersion,
			"max_upload_size":       vhost.MaxUploadSize,
			"proxy_read_timeout":    vhost.ProxyReadTimeout,
			"proxy_connect_timeout": vhost.ProxyConnectTimeout,
			"bot_detection_enabled": vhost.BotDetectionEnabled,
			"bot_detection_type":    vhost.BotDetectionType,
			"recaptcha_version":     vhost.RecaptchaVersion,
			"rate_limit_enabled":    vhost.RateLimitEnabled,
			"rate_limit_requests":   vhost.RateLimitRequests,
			"rate_limit_window":     vhost.RateLimitWindow,
			"defense_mode":          vhost.DefenseMode,
			"custom_headers":        vhost.CustomHeaders,
			"custom_locations":      customLocs,
			"created_at":            vhost.CreatedAt,
			"updated_at":            vhost.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, response)
}

// GetVHost returns a specific virtual host
func (h *VHostHandler) GetVHost(c *gin.Context) {
	type VHost struct {
		ID                     string          `db:"id" json:"id"`
		Name                   string          `db:"name" json:"name"`
		Type                   string          `db:"type" json:"type"`
		Domain                 string          `db:"domain" json:"domain"`
		BackendURL             string          `db:"backend_url" json:"backend_url"`
		Backends               *string         `db:"backends" json:"backends"`
		LoadBalanceMethod      *string         `db:"load_balance_method" json:"load_balance_method"`
		CustomConfig           *string         `db:"custom_config" json:"custom_config"`
		SSLEnabled             bool            `db:"ssl_enabled" json:"ssl_enabled"`
		SSLCertificateID       *string         `db:"ssl_certificate_id" json:"ssl_certificate_id"`
		SSLCertPath            *string         `db:"ssl_cert_path" json:"ssl_cert_path"`
		SSLKeyPath             *string         `db:"ssl_key_path" json:"ssl_key_path"`
		Enabled                bool            `db:"enabled" json:"enabled"`
		WebsocketEnabled       bool            `db:"websocket_enabled" json:"websocket_enabled"`
		HTTPVersion            string          `db:"http_version" json:"http_version"`
		TLSVersion             string          `db:"tls_version" json:"tls_version"`
		MaxUploadSize          int             `db:"max_upload_size" json:"max_upload_size"`
		ProxyReadTimeout       int             `db:"proxy_read_timeout" json:"proxy_read_timeout"`
		ProxyConnectTimeout    int             `db:"proxy_connect_timeout" json:"proxy_connect_timeout"`
		BotDetectionEnabled    bool            `db:"bot_detection_enabled" json:"bot_detection_enabled"`
		BotDetectionType       string          `db:"bot_detection_type" json:"bot_detection_type"`
		RecaptchaVersion       string          `db:"recaptcha_version" json:"recaptcha_version"`
		RateLimitEnabled       bool            `db:"rate_limit_enabled" json:"rate_limit_enabled"`
		RateLimitRequests      int             `db:"rate_limit_requests" json:"rate_limit_requests"`
		RateLimitWindow        int             `db:"rate_limit_window" json:"rate_limit_window"`
		RegionFilteringEnabled bool            `db:"region_filtering_enabled" json:"region_filtering_enabled"`
		RegionWhitelist        json.RawMessage `db:"region_whitelist" json:"region_whitelist"`
		RegionBlacklist        json.RawMessage `db:"region_blacklist" json:"region_blacklist"`
		DefenseMode            string          `db:"defense_mode" json:"defense_mode"`
		CustomHeaders          json.RawMessage `db:"custom_headers" json:"custom_headers"`
		CreatedAt              time.Time       `db:"created_at" json:"created_at"`
		UpdatedAt              time.Time       `db:"updated_at" json:"updated_at"`
	}

	id := c.Param("id")

	var vhost VHost
	query := `
		SELECT id::text, name, COALESCE(type, 'proxy') as type, domain, backend_url,
		       backends::text as backends, COALESCE(load_balance_method, 'round_robin') as load_balance_method, custom_config,
		       ssl_enabled, ssl_certificate_id::text, ssl_cert_path, ssl_key_path, enabled,
		       websocket_enabled, http_version, tls_version, max_upload_size,
		       proxy_read_timeout, proxy_connect_timeout,
		       bot_detection_enabled, bot_detection_type, recaptcha_version,
		       rate_limit_enabled, rate_limit_requests, rate_limit_window,
		       region_filtering_enabled, region_whitelist, region_blacklist,
		       COALESCE(defense_mode, 'defense') as defense_mode, custom_headers, created_at, updated_at
		FROM vhosts 
		WHERE id = $1
	`

	err := h.db.Get(&vhost, query, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrVHostNotFound})
		return
	}

	// Parse fields
	backends := parseJSONStringArray(vhost.Backends)
	regionWhitelist := parseJSONRawMessage(vhost.RegionWhitelist)
	regionBlacklist := parseJSONRawMessage(vhost.RegionBlacklist)

	loadBalanceMethod := constants.DefaultLoadBalanceMethod
	if vhost.LoadBalanceMethod != nil {
		loadBalanceMethod = *vhost.LoadBalanceMethod
	}

	customConfig := ""
	if vhost.CustomConfig != nil {
		customConfig = *vhost.CustomConfig
	}

	// Query custom locations
	type CustomLocation struct {
		Path              string  `db:"path" json:"path"`
		ProxyPass         *string `db:"proxy_pass" json:"proxy_pass"`
		CustomConfig      *string `db:"custom_config" json:"config"`
		WebSocketEnabled  bool    `db:"websocket_enabled" json:"websocket_enabled"`
		Backends          *string `db:"backends" json:"backends"`
		LoadBalanceMethod *string `db:"load_balance_method" json:"load_balance_method"`
	}
	var dbCustomLocations []CustomLocation
	locQuery := `
		SELECT path, proxy_pass, custom_config, COALESCE(websocket_enabled, false) as websocket_enabled,
		       backends::text as backends, COALESCE(load_balance_method, 'round_robin') as load_balance_method
		FROM vhost_locations
		WHERE vhost_id = $1 AND enabled = true
		ORDER BY created_at ASC
	`
	_ = h.db.Select(&dbCustomLocations, locQuery, id)

	// Process custom locations to parse JSON backends
	var customLocations []map[string]interface{}
	for _, loc := range dbCustomLocations {
		var locBackends []string
		if loc.Backends != nil && *loc.Backends != "" && *loc.Backends != "[]" {
			_ = json.Unmarshal([]byte(*loc.Backends), &locBackends)
		}
		if locBackends == nil {
			locBackends = []string{}
		}

		locLoadBalanceMethod := "round_robin"
		if loc.LoadBalanceMethod != nil {
			locLoadBalanceMethod = *loc.LoadBalanceMethod
		}

		customLocations = append(customLocations, map[string]interface{}{
			"path":                loc.Path,
			"proxy_pass":          loc.ProxyPass,
			"config":              loc.CustomConfig,
			"websocket_enabled":   loc.WebSocketEnabled,
			"backends":            locBackends,
			"load_balance_method": locLoadBalanceMethod,
		})
	}

	// Build response with all fields
	response := map[string]interface{}{
		"id":                       vhost.ID,
		"name":                     vhost.Name,
		"type":                     vhost.Type,
		"domain":                   vhost.Domain,
		"backend_url":              vhost.BackendURL,
		"backends":                 backends,
		"load_balance_method":      loadBalanceMethod,
		"custom_config":            customConfig,
		"ssl_enabled":              vhost.SSLEnabled,
		"ssl_certificate_id":       vhost.SSLCertificateID,
		"ssl_cert_path":            vhost.SSLCertPath,
		"ssl_key_path":             vhost.SSLKeyPath,
		"enabled":                  vhost.Enabled,
		"websocket_enabled":        vhost.WebsocketEnabled,
		"http_version":             vhost.HTTPVersion,
		"tls_version":              vhost.TLSVersion,
		"max_upload_size":          vhost.MaxUploadSize,
		"proxy_read_timeout":       vhost.ProxyReadTimeout,
		"proxy_connect_timeout":    vhost.ProxyConnectTimeout,
		"bot_detection_enabled":    vhost.BotDetectionEnabled,
		"bot_detection_type":       vhost.BotDetectionType,
		"recaptcha_version":        vhost.RecaptchaVersion,
		"rate_limit_enabled":       vhost.RateLimitEnabled,
		"rate_limit_requests":      vhost.RateLimitRequests,
		"rate_limit_window":        vhost.RateLimitWindow,
		"region_filtering_enabled": vhost.RegionFilteringEnabled,
		"region_whitelist":         regionWhitelist,
		"region_blacklist":         regionBlacklist,
		"custom_headers":           vhost.CustomHeaders,
		"custom_locations":         customLocations,
		"created_at":               vhost.CreatedAt,
		"updated_at":               vhost.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// CreateVHost creates a new virtual host
func (h *VHostHandler) CreateVHost(c *gin.Context) {
	var input models.VHostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sanitizeVHostInput(&input)

	// Marshal custom_headers to JSON
	customHeadersJSON, err := json.Marshal(input.CustomHeaders)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid custom headers format"})
		return
	}

	// Marshal backends to JSON
	backendsJSON, err := json.Marshal(input.Backends)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backends format"})
		return
	}

	// Validate SSL Certificate if provided
	if input.SSLCertificateID != nil {
		// Verify certificate exists
		_, err := h.certService.GetCertificate(*input.SSLCertificateID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SSL certificate ID"})
			return
		}
	}

	query := `
		INSERT INTO vhosts (id, name, type, domain, backend_url, backends, load_balance_method, custom_config,
		                   ssl_enabled, ssl_certificate_id, ssl_cert_path, ssl_key_path, enabled,
		                   websocket_enabled, http_version, tls_version, max_upload_size,
		                   proxy_read_timeout, proxy_connect_timeout,
		                   bot_detection_enabled, bot_detection_type, recaptcha_version,
		                   rate_limit_enabled, rate_limit_requests, rate_limit_window,
		                   region_whitelist, region_blacklist, region_filtering_enabled,
		                   defense_mode, custom_headers, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31)
		RETURNING id
	`

	var id string
	err = h.db.QueryRow(query,
		input.Name,
		input.Type,
		input.Domain,
		input.BackendURL,
		backendsJSON,
		input.LoadBalanceMethod,
		input.CustomConfig,
		input.SSLEnabled,
		input.SSLCertificateID,
		input.SSLCertPath,
		input.SSLKeyPath,
		input.Enabled,
		input.WebsocketEnabled,
		input.HTTPVersion,
		input.TLSVersion,
		input.MaxUploadSize,
		input.ProxyReadTimeout,
		input.ProxyConnectTimeout,
		input.BotDetectionEnabled,
		input.BotDetectionType,
		input.RecaptchaVersion,
		input.RateLimitEnabled,
		input.RateLimitRequests,
		input.RateLimitWindow,
		pq.Array(input.RegionWhitelist),
		pq.Array(input.RegionBlacklist),
		input.RegionFilteringEnabled,
		customHeadersJSON,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Insert custom locations if any
	if len(input.CustomLocations) > 0 {
		for _, loc := range input.CustomLocations {
			locQuery := `
				INSERT INTO vhost_locations (id, vhost_id, path, backend_url, proxy_pass, custom_config, websocket_enabled, backends, load_balance_method, enabled, created_at, updated_at)
				VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, true, $9, $10)
			`
			proxyPass := loc["proxy_pass"]
			websocketEnabled := false
			if wsEnabled, ok := loc["websocket_enabled"].(bool); ok {
				websocketEnabled = wsEnabled
			}

			// Handle backends for location
			var backendsJSON []byte = []byte("[]")
			if backends, ok := loc["backends"].([]interface{}); ok {
				if b, err := json.Marshal(backends); err == nil {
					backendsJSON = b
				}
			}

			loadBalanceMethod := "round_robin"
			if method, ok := loc["load_balance_method"].(string); ok && method != "" {
				loadBalanceMethod = method
			}

			_, err := h.db.Exec(locQuery,
				id,
				loc["path"],
				proxyPass,         // backend_url (required NOT NULL)
				proxyPass,         // proxy_pass
				loc["config"],     // custom_config
				websocketEnabled,  // websocket_enabled
				backendsJSON,      // backends
				loadBalanceMethod, // load_balance_method
				time.Now(),
				time.Now(),
			)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create location: " + err.Error()})
				return
			}
		}
	}

	// Get the created vhost
	vhost, err := h.vhostService.GetVHostByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve created vhost"})
		return
	}

	// If SSL is enabled and certificate ID is provided, save certificate files
	if input.SSLEnabled && input.SSLCertificateID != nil && *input.SSLCertificateID != "" {
		cert, err := h.certService.GetCertificate(*input.SSLCertificateID)
		if err == nil {
			// Save certificate files to filesystem
			h.certService.SaveCertificateFiles(*input.SSLCertificateID, []byte(cert.CertContent), []byte(cert.KeyContent))
		}
	}

	// Generate nginx config for this vhost
	if err := h.nginxConfigService.GenerateVHostConfig(vhost); err != nil {
		fmt.Printf("Warning: Failed to generate nginx config: %v\n", err)
	}

	// Reload nginx
	h.reloadNginx()

	// Reload proxy map to include new vhost
	if h.proxyReloader != nil {
		if err := h.proxyReloader.ReloadVHosts(); err != nil {
			fmt.Printf(proxyReloadWarningMsg, err)
		}
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "VHost created successfully"})
}

// UpdateVHost updates an existing virtual host
func (h *VHostHandler) UpdateVHost(c *gin.Context) {
	id := c.Param("id")

	var input models.VHostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sanitizeVHostInput(&input)

	// Validate SSL Certificate if provided
	// Validate SSL Certificate if provided
	if input.SSLCertificateID != nil {
		// Verify certificate exists
		_, err := h.certService.GetCertificate(*input.SSLCertificateID)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid SSL certificate ID"})
			return
		}
	}

	query := `
		UPDATE vhosts 
		SET name = $1, type = $2, domain = $3, backend_url = $4, backends = $5, load_balance_method = $6, custom_config = $7,
		    ssl_enabled = $8, ssl_certificate_id = $9, ssl_cert_path = $10, ssl_key_path = $11, enabled = $12,
		    websocket_enabled = $13, http_version = $14, tls_version = $15, max_upload_size = $16,
		    proxy_read_timeout = $17, proxy_connect_timeout = $18,
		    bot_detection_enabled = $19, bot_detection_type = $20, recaptcha_version = $21,
		    rate_limit_enabled = $22, rate_limit_requests = $23, rate_limit_window = $24,
		    region_whitelist = $25, region_blacklist = $26, region_filtering_enabled = $27,
		    defense_mode = $28, custom_headers = $29, updated_at = $30
		WHERE id = $31
	`

	// Marshal custom_headers to JSON
	customHeadersJSON, err := json.Marshal(input.CustomHeaders)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid custom headers format"})
		return
	}

	// Marshal backends to JSON
	backendsJSON, err := json.Marshal(input.Backends)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backends format"})
		return
	}

	_, err = h.db.Exec(query,
		input.Name,
		input.Type,
		input.Domain,
		input.BackendURL,
		backendsJSON,
		input.LoadBalanceMethod,
		input.CustomConfig,
		input.SSLEnabled,
		input.SSLCertificateID,
		input.SSLCertPath,
		input.SSLKeyPath,
		input.Enabled,
		input.WebsocketEnabled,
		input.HTTPVersion,
		input.TLSVersion,
		input.MaxUploadSize,
		input.ProxyReadTimeout,
		input.ProxyConnectTimeout,
		input.BotDetectionEnabled,
		input.BotDetectionType,
		input.RecaptchaVersion,
		input.RateLimitEnabled,
		input.RateLimitRequests,
		input.RateLimitWindow,
		pq.Array(input.RegionWhitelist),
		pq.Array(input.RegionBlacklist),
		input.RegionFilteringEnabled,
		customHeadersJSON,
		time.Now(),
		id,
	)

	if err != nil {
		fmt.Printf("Error updating vhost: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Delete existing custom locations and insert new ones
	_, err = h.db.Exec("DELETE FROM vhost_locations WHERE vhost_id = $1", id)
	if err != nil {
		fmt.Printf("Error deleting old locations: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete old locations: " + err.Error()})
		return
	}

	// Insert new custom locations if any
	if len(input.CustomLocations) > 0 {
		for _, loc := range input.CustomLocations {
			locQuery := `
				INSERT INTO vhost_locations (id, vhost_id, path, backend_url, proxy_pass, custom_config, websocket_enabled, backends, load_balance_method, enabled, created_at, updated_at)
				VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, true, $9, $10)
			`
			proxyPass := loc["proxy_pass"]
			websocketEnabled := false
			if wsEnabled, ok := loc["websocket_enabled"].(bool); ok {
				websocketEnabled = wsEnabled
			}

			// Handle backends for location
			var backendsJSON []byte = []byte("[]")
			if backends, ok := loc["backends"].([]interface{}); ok {
				if b, err := json.Marshal(backends); err == nil {
					backendsJSON = b
				}
			}

			loadBalanceMethod := "round_robin"
			if method, ok := loc["load_balance_method"].(string); ok && method != "" {
				loadBalanceMethod = method
			}

			_, err := h.db.Exec(locQuery,
				id,
				loc["path"],
				proxyPass,         // backend_url (required NOT NULL)
				proxyPass,         // proxy_pass
				loc["config"],     // custom_config
				websocketEnabled,  // websocket_enabled
				backendsJSON,      // backends
				loadBalanceMethod, // load_balance_method
				time.Now(),
				time.Now(),
			)
			if err != nil {
				fmt.Printf("Error creating location in UpdateVHost: %v, loc data: %+v\n", err, loc)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create location: " + err.Error()})
				return
			}
		}
	}

	// Get the updated vhost
	vhost, err := h.vhostService.GetVHostByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve updated vhost"})
		return
	}

	// If SSL is enabled and certificate ID is provided, save certificate files
	if input.SSLEnabled && input.SSLCertificateID != nil && *input.SSLCertificateID != "" {
		cert, err := h.certService.GetCertificate(*input.SSLCertificateID)
		if err == nil {
			// Save certificate files to filesystem
			h.certService.SaveCertificateFiles(*input.SSLCertificateID, []byte(cert.CertContent), []byte(cert.KeyContent))
		}
	}

	// Regenerate nginx config for this vhost
	if err := h.nginxConfigService.GenerateVHostConfig(vhost); err != nil {
		fmt.Printf("Warning: Failed to generate nginx config: %v\n", err)
	}

	// Reload nginx
	h.reloadNginx()

	// Reload proxy map
	if h.proxyReloader != nil {
		if err := h.proxyReloader.ReloadVHosts(); err != nil {
			fmt.Printf(proxyReloadWarningMsg, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "VHost updated successfully"})
}

// PreviewVHostConfig generates a preview of the nginx configuration without saving
func (h *VHostHandler) PreviewVHostConfig(c *gin.Context) {
	var input models.VHostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sanitizeVHostInput(&input)

	// Construct VHostWithLocations for preview
	// Note: We are using services.VHostWithLocations struct structure manually
	// because we don't have a direct model -> VHostWithLocations converter that accepts input struct directly

	// First, map input to VHost model
	vhost := &services.VHostWithLocations{
		VHost: &models.VHost{
			Name:                   input.Name,
			Type:                   input.Type,
			Domain:                 input.Domain,
			BackendURL:             input.BackendURL,
			Backends:               input.Backends,
			LoadBalanceMethod:      input.LoadBalanceMethod,
			CustomConfig:           input.CustomConfig,
			SSLEnabled:             input.SSLEnabled,
			SSLCertificateID:       "", // Not needed for preview usually
			Enabled:                input.Enabled,
			RegionWhitelist:        input.RegionWhitelist,
			RegionBlacklist:        input.RegionBlacklist,
			RegionFilteringEnabled: input.RegionFilteringEnabled,
			DefenseMode:            input.DefenseMode,
		},
		CustomLocations: []services.CustomLocation{},
	}

	if input.SSLCertificateID != nil {
		vhost.SSLCertificateID = *input.SSLCertificateID
	}

	// Helper to sanitize upstream name
	sanitize := func(s string) string {
		re := regexp.MustCompile(`[^a-zA-Z0-9]`)
		return re.ReplaceAllString(s, "_")
	}

	// Determine upstream name
	upstreamName := sanitize(input.Domain)
	vhost.UpstreamName = upstreamName

	// If there are additional backends, prepend backend_url to form the full upstream
	allBackends := input.Backends
	if len(allBackends) > 0 && input.BackendURL != "" {
		allBackends = append([]string{input.BackendURL}, allBackends...)
	}
	if len(allBackends) > 0 {
		vhost.HasUpstream = true
		vhost.Backends = allBackends
	}

	// Process Custom Locations
	for _, loc := range input.CustomLocations {
		path, _ := loc["path"].(string)
		proxyPass, _ := loc["proxy_pass"].(string)
		config, _ := loc["config"].(string)
		wsEnabled, _ := loc["websocket_enabled"].(bool)

		customLoc := services.CustomLocation{
			Path:             path,
			ProxyPass:        proxyPass,
			CustomConfig:     config,
			WebSocketEnabled: wsEnabled,
		}

		// Handle backends for location
		if backends, ok := loc["backends"].([]interface{}); ok {
			var locBackends []string
			// Convert []interface{} to []string
			for _, b := range backends {
				if s, ok := b.(string); ok {
					locBackends = append(locBackends, s)
				}
			}

			if len(locBackends) > 0 {
				customLoc.Backends = locBackends
				customLoc.HasUpstream = true
				// sanitize path: remove leading slash first
				cleanPath := strings.TrimPrefix(path, "/")
				customLoc.UpstreamName = upstreamName + "_loc_" + sanitize(cleanPath)
			}
		}

		if method, ok := loc["load_balance_method"].(string); ok && method != "" {
			customLoc.LoadBalanceMethod = method
		}

		vhost.CustomLocations = append(vhost.CustomLocations, customLoc)
	}

	// Generate config
	configContent, err := h.nginxConfigService.GenerateVHostConfigContent(vhost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate config preview: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"config": configContent})
}

// DeleteVHost deletes a virtual host
func (h *VHostHandler) DeleteVHost(c *gin.Context) {
	id := c.Param("id")

	// Get vhost info before deletion
	vhost, err := h.vhostService.GetVHostByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrVHostNotFound})
		return
	}

	// Delete the vhost (CASCADE will handle related records)
	result, err := h.db.Exec("DELETE FROM vhosts WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vhost: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrVHostNotFound})
		return
	}

	// Delete nginx config
	if err := h.nginxConfigService.DeleteVHostConfig(vhost.Domain); err != nil {
		fmt.Printf("Warning: Failed to delete nginx config: %v\n", err)
	}

	// Delete log files
	logDir := "/data/nginx/logs"
	accessLogPath := fmt.Sprintf("%s/%s_access.log", logDir, vhost.Domain)
	errorLogPath := fmt.Sprintf("%s/%s_error.log", logDir, vhost.Domain)

	if err := os.Remove(accessLogPath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: Failed to delete access log %s: %v\n", accessLogPath, err)
	}

	if err := os.Remove(errorLogPath); err != nil && !os.IsNotExist(err) {
		fmt.Printf("Warning: Failed to delete error log %s: %v\n", errorLogPath, err)
	}

	// Reload nginx
	h.reloadNginx()

	// Reload proxy map
	if h.proxyReloader != nil {
		if err := h.proxyReloader.ReloadVHosts(); err != nil {
			fmt.Printf(proxyReloadWarningMsg, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "VHost deleted successfully"})
}

// GetVHostConfig returns the nginx config content for a vhost
func (h *VHostHandler) GetVHostConfig(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is required"})
		return
	}

	configPath := fmt.Sprintf("/data/nginx/config/%s.conf", domain)
	content, err := os.ReadFile(configPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config file not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domain":  domain,
		"content": string(content),
		"path":    configPath,
	})
}

// UpdateVHostConfig updates the nginx config content for a vhost
func (h *VHostHandler) UpdateVHostConfig(c *gin.Context) {
	domain := c.Param("domain")
	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Domain is required"})
		return
	}

	var req struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	configPath := fmt.Sprintf("/data/nginx/config/%s.conf", domain)

	// Backup existing config
	backupPath := configPath + ".backup"
	if existingContent, err := os.ReadFile(configPath); err == nil {
		if err := os.WriteFile(backupPath, existingContent, 0644); err != nil {
			fmt.Printf("Warning: Failed to create backup: %v\n", err)
		}
	}

	// Write new config
	if err := os.WriteFile(configPath, []byte(req.Content), 0644); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to write config file"})
		return
	}

	// Reload nginx
	h.reloadNginx()

	c.JSON(http.StatusOK, gin.H{
		"message": "Config updated successfully",
		"domain":  domain,
		"backup":  backupPath,
	})
}

// reloadNginx sends reload signal to nginx
func (h *VHostHandler) reloadNginx() {
	// Instead of using docker exec, we'll create a reload signal file
	// A separate script or nginx itself can watch this file
	signalFile := "/data/nginx/.reload"
	if err := os.WriteFile(signalFile, []byte(time.Now().Format(time.RFC3339)), 0644); err != nil {
		fmt.Printf("Warning: Failed to create reload signal: %v\n", err)
	}
	fmt.Println("Nginx reload signal created, manual reload may be needed: docker compose exec nginx-proxy nginx -s reload")
}

// RegenerateAllConfigs regenerates nginx config files for all vhosts
func (h *VHostHandler) RegenerateAllConfigs(c *gin.Context) {
	// Get all enabled vhosts
	vhosts, err := h.vhostService.ListVHosts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vhosts: " + err.Error()})
		return
	}

	var regenerated []string
	var errors []string

	for _, vhost := range vhosts {
		if !vhost.Enabled {
			continue
		}

		// Generate nginx config
		if err := h.nginxConfigService.GenerateVHostConfig(vhost); err != nil {
			errors = append(errors, fmt.Sprintf("%s: %v", vhost.Domain, err))
			continue
		}
		regenerated = append(regenerated, vhost.Domain)
	}

	// Reload nginx
	h.reloadNginx()

	// Reload proxy map
	if h.proxyReloader != nil {
		if err := h.proxyReloader.ReloadVHosts(); err != nil {
			fmt.Printf(proxyReloadWarningMsg, err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Configs regenerated successfully",
		"regenerated": regenerated,
		"count":       len(regenerated),
		"errors":      errors,
	})
}
