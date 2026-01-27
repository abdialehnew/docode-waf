package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/aleh/docode-waf/internal/constants"
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
		CustomHeaders       json.RawMessage `db:"custom_headers" json:"custom_headers"`
		CreatedAt           time.Time       `db:"created_at" json:"created_at"`
		UpdatedAt           time.Time       `db:"updated_at" json:"updated_at"`
	}

	var vhosts []VHost

	query := `
		SELECT id::text, name, domain, backend_url, 
		       backends::text as backends, COALESCE(load_balance_method, 'round_robin') as load_balance_method, custom_config,
		       ssl_enabled, ssl_certificate_id::text, ssl_cert_path, ssl_key_path, enabled,
		       websocket_enabled, http_version, tls_version, max_upload_size,
		       proxy_read_timeout, proxy_connect_timeout,
		       bot_detection_enabled, bot_detection_type, recaptcha_version,
		       rate_limit_enabled, rate_limit_requests, rate_limit_window,
		       custom_headers, created_at, updated_at
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
		VHostID          string  `db:"vhost_id"`
		Path             string  `db:"path" json:"path"`
		ProxyPass        *string `db:"proxy_pass" json:"proxy_pass"`
		CustomConfig     *string `db:"custom_config" json:"config"`
		WebSocketEnabled bool    `db:"websocket_enabled" json:"websocket_enabled"`
	}
	var allLocations []CustomLocation
	locQuery := `
		SELECT vhost_id::text, path, proxy_pass, custom_config, COALESCE(websocket_enabled, false) as websocket_enabled
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
		locationMap[loc.VHostID] = append(locationMap[loc.VHostID], map[string]interface{}{
			"path":              loc.Path,
			"proxy_pass":        loc.ProxyPass,
			"config":            loc.CustomConfig,
			"websocket_enabled": loc.WebSocketEnabled,
		})
	}

	// Build response with custom_locations
	var response []map[string]interface{}
	for _, vhost := range vhosts {
		customLocs := locationMap[vhost.ID]
		if customLocs == nil {
			customLocs = []map[string]interface{}{}
		}

		// Parse backends from JSON
		var backends []string
		if vhost.Backends != nil && *vhost.Backends != "" && *vhost.Backends != "[]" {
			_ = json.Unmarshal([]byte(*vhost.Backends), &backends)
		}
		if backends == nil {
			backends = []string{}
		}

		loadBalanceMethod := "round_robin"
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
		ID                  string          `db:"id" json:"id"`
		Name                string          `db:"name" json:"name"`
		Domain              string          `db:"domain" json:"domain"`
		BackendURL          string          `db:"backend_url" json:"backend_url"`
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
		CustomHeaders       json.RawMessage `db:"custom_headers" json:"custom_headers"`
		CreatedAt           time.Time       `db:"created_at" json:"created_at"`
		UpdatedAt           time.Time       `db:"updated_at" json:"updated_at"`
	}

	id := c.Param("id")

	var vhost VHost
	query := `
		SELECT id::text, name, domain, backend_url, ssl_enabled, 
		       ssl_certificate_id::text, ssl_cert_path, ssl_key_path, enabled,
		       websocket_enabled, http_version, tls_version, max_upload_size,
		       proxy_read_timeout, proxy_connect_timeout, custom_headers,
		       created_at, updated_at
		FROM vhosts 
		WHERE id = $1
	`

	err := h.db.Get(&vhost, query, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrVHostNotFound})
		return
	}

	// Query custom locations
	type CustomLocation struct {
		Path             string  `db:"path" json:"path"`
		ProxyPass        *string `db:"proxy_pass" json:"proxy_pass"`
		CustomConfig     *string `db:"custom_config" json:"config"`
		WebSocketEnabled bool    `db:"websocket_enabled" json:"websocket_enabled"`
	}
	var customLocations []CustomLocation
	locQuery := `
		SELECT path, proxy_pass, custom_config, COALESCE(websocket_enabled, false) as websocket_enabled
		FROM vhost_locations
		WHERE vhost_id = $1 AND enabled = true
		ORDER BY created_at ASC
	`
	_ = h.db.Select(&customLocations, locQuery, id)

	// Build response with custom_locations
	response := map[string]interface{}{
		"id":                    vhost.ID,
		"name":                  vhost.Name,
		"domain":                vhost.Domain,
		"backend_url":           vhost.BackendURL,
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
		"custom_headers":        vhost.CustomHeaders,
		"custom_locations":      customLocations,
		"created_at":            vhost.CreatedAt,
		"updated_at":            vhost.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// CreateVHost creates a new virtual host
func (h *VHostHandler) CreateVHost(c *gin.Context) {
	var input struct {
		Name                   string                   `json:"name" binding:"required"`
		Domain                 string                   `json:"domain" binding:"required"`
		BackendURL             string                   `json:"backend_url" binding:"required"`
		Backends               []string                 `json:"backends"`
		LoadBalanceMethod      string                   `json:"load_balance_method"`
		CustomConfig           string                   `json:"custom_config"`
		SSLEnabled             bool                     `json:"ssl_enabled"`
		SSLCertificateID       *string                  `json:"ssl_certificate_id"`
		SSLCertPath            string                   `json:"ssl_cert_path"`
		SSLKeyPath             string                   `json:"ssl_key_path"`
		Enabled                bool                     `json:"enabled"`
		WebsocketEnabled       bool                     `json:"websocket_enabled"`
		HTTPVersion            string                   `json:"http_version"`
		TLSVersion             string                   `json:"tls_version"`
		MaxUploadSize          int                      `json:"max_upload_size"`
		ProxyReadTimeout       int                      `json:"proxy_read_timeout"`
		ProxyConnectTimeout    int                      `json:"proxy_connect_timeout"`
		BotDetectionEnabled    bool                     `json:"bot_detection_enabled"`
		BotDetectionType       string                   `json:"bot_detection_type"`
		RecaptchaVersion       string                   `json:"recaptcha_version"`
		RateLimitEnabled       bool                     `json:"rate_limit_enabled"`
		RateLimitRequests      int                      `json:"rate_limit_requests"`
		RateLimitWindow        int                      `json:"rate_limit_window"`
		RegionWhitelist        []string                 `json:"region_whitelist"`
		RegionBlacklist        []string                 `json:"region_blacklist"`
		RegionFilteringEnabled bool                     `json:"region_filtering_enabled"`
		CustomHeaders          map[string]interface{}   `json:"custom_headers"`
		CustomLocations        []map[string]interface{} `json:"custom_locations"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if input.HTTPVersion == "" {
		input.HTTPVersion = "http/1.1"
	}
	if input.ProxyReadTimeout == 0 {
		input.ProxyReadTimeout = 60
	}
	if input.ProxyConnectTimeout == 0 {
		input.ProxyConnectTimeout = 60
	}
	if input.BotDetectionType == "" {
		input.BotDetectionType = "turnstile"
	}
	if input.RecaptchaVersion == "" {
		input.RecaptchaVersion = "v2"
	}
	if input.RateLimitRequests == 0 {
		input.RateLimitRequests = 100
	}
	if input.RateLimitWindow == 0 {
		input.RateLimitWindow = 60
	}
	if input.CustomHeaders == nil {
		input.CustomHeaders = make(map[string]interface{})
	}

	// Marshal custom_headers to JSON
	customHeadersJSON, err := json.Marshal(input.CustomHeaders)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid custom headers format"})
		return
	}

	// Marshal backends to JSON
	if input.Backends == nil {
		input.Backends = []string{}
	}
	backendsJSON, err := json.Marshal(input.Backends)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backends format"})
		return
	}

	// Set default load balance method
	if input.LoadBalanceMethod == "" {
		input.LoadBalanceMethod = "round_robin"
	}

	// Sanitize SSL Certificate ID
	if input.SSLCertificateID != nil && *input.SSLCertificateID == "" {
		input.SSLCertificateID = nil
	}

	query := `
		INSERT INTO vhosts (id, name, domain, backend_url, backends, load_balance_method, custom_config,
		                   ssl_enabled, ssl_certificate_id, ssl_cert_path, ssl_key_path, enabled,
		                   websocket_enabled, http_version, tls_version, max_upload_size,
		                   proxy_read_timeout, proxy_connect_timeout,
		                   bot_detection_enabled, bot_detection_type, recaptcha_version,
		                   rate_limit_enabled, rate_limit_requests, rate_limit_window,
		                   region_whitelist, region_blacklist, region_filtering_enabled,
		                   custom_headers, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29)
		RETURNING id
	`

	var id string
	err = h.db.QueryRow(query,
		input.Name,
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
				INSERT INTO vhost_locations (id, vhost_id, path, backend_url, proxy_pass, custom_config, websocket_enabled, enabled, created_at, updated_at)
				VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, true, $7, $8)
			`
			proxyPass := loc["proxy_pass"]
			websocketEnabled := false
			if wsEnabled, ok := loc["websocket_enabled"].(bool); ok {
				websocketEnabled = wsEnabled
			}
			_, err := h.db.Exec(locQuery,
				id,
				loc["path"],
				proxyPass,        // backend_url (required NOT NULL)
				proxyPass,        // proxy_pass
				loc["config"],    // custom_config
				websocketEnabled, // websocket_enabled
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

	var input struct {
		Name                   string                   `json:"name"`
		Domain                 string                   `json:"domain"`
		BackendURL             string                   `json:"backend_url"`
		Backends               []string                 `json:"backends"`
		LoadBalanceMethod      string                   `json:"load_balance_method"`
		CustomConfig           string                   `json:"custom_config"`
		SSLEnabled             bool                     `json:"ssl_enabled"`
		SSLCertificateID       *string                  `json:"ssl_certificate_id"`
		SSLCertPath            string                   `json:"ssl_cert_path"`
		SSLKeyPath             string                   `json:"ssl_key_path"`
		Enabled                bool                     `json:"enabled"`
		WebsocketEnabled       bool                     `json:"websocket_enabled"`
		HTTPVersion            string                   `json:"http_version"`
		TLSVersion             string                   `json:"tls_version"`
		MaxUploadSize          int                      `json:"max_upload_size"`
		ProxyReadTimeout       int                      `json:"proxy_read_timeout"`
		ProxyConnectTimeout    int                      `json:"proxy_connect_timeout"`
		BotDetectionEnabled    bool                     `json:"bot_detection_enabled"`
		BotDetectionType       string                   `json:"bot_detection_type"`
		RecaptchaVersion       string                   `json:"recaptcha_version"`
		RateLimitEnabled       bool                     `json:"rate_limit_enabled"`
		RateLimitRequests      int                      `json:"rate_limit_requests"`
		RateLimitWindow        int                      `json:"rate_limit_window"`
		RegionWhitelist        []string                 `json:"region_whitelist"`
		RegionBlacklist        []string                 `json:"region_blacklist"`
		RegionFilteringEnabled bool                     `json:"region_filtering_enabled"`
		CustomHeaders          map[string]interface{}   `json:"custom_headers"`
		CustomLocations        []map[string]interface{} `json:"custom_locations"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		UPDATE vhosts 
		SET name = $1, domain = $2, backend_url = $3, backends = $4, load_balance_method = $5, custom_config = $6,
		    ssl_enabled = $7, ssl_certificate_id = $8, ssl_cert_path = $9, ssl_key_path = $10, enabled = $11,
		    websocket_enabled = $12, http_version = $13, tls_version = $14, max_upload_size = $15,
		    proxy_read_timeout = $16, proxy_connect_timeout = $17,
		    bot_detection_enabled = $18, bot_detection_type = $19, recaptcha_version = $20,
		    rate_limit_enabled = $21, rate_limit_requests = $22, rate_limit_window = $23,
		    region_whitelist = $24, region_blacklist = $25, region_filtering_enabled = $26,
		    custom_headers = $27, updated_at = $28
		WHERE id = $29
	`

	// Set defaults
	if input.HTTPVersion == "" {
		input.HTTPVersion = "http/1.1"
	}
	if input.MaxUploadSize == 0 {
		input.MaxUploadSize = 10
	}
	if input.ProxyReadTimeout == 0 {
		input.ProxyReadTimeout = 60
	}
	if input.ProxyConnectTimeout == 0 {
		input.ProxyConnectTimeout = 60
	}
	if input.CustomHeaders == nil {
		input.CustomHeaders = make(map[string]interface{})
	}
	if input.Backends == nil {
		input.Backends = []string{}
	}

	// Sanitize SSL Certificate ID
	if input.SSLCertificateID != nil && *input.SSLCertificateID == "" {
		input.SSLCertificateID = nil
	}
	if input.LoadBalanceMethod == "" {
		input.LoadBalanceMethod = "round_robin"
	}

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
				INSERT INTO vhost_locations (id, vhost_id, path, backend_url, proxy_pass, custom_config, websocket_enabled, enabled, created_at, updated_at)
				VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, true, $7, $8)
			`
			proxyPass := loc["proxy_pass"]
			websocketEnabled := false
			if wsEnabled, ok := loc["websocket_enabled"].(bool); ok {
				websocketEnabled = wsEnabled
			}
			_, err := h.db.Exec(locQuery,
				id,
				loc["path"],
				proxyPass,        // backend_url (required NOT NULL)
				proxyPass,        // proxy_pass
				loc["config"],    // custom_config
				websocketEnabled, // websocket_enabled
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
