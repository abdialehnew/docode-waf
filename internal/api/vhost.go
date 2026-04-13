package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
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

// vhostDBResponse is a internal struct for database mapping
type vhostDBResponse struct {
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
	RegionWhitelist        pq.StringArray  `db:"region_whitelist" json:"region_whitelist"`
	RegionBlacklist        pq.StringArray  `db:"region_blacklist" json:"region_blacklist"`
	DefenseMode            string          `db:"defense_mode" json:"defense_mode"`
	CustomHeaders          json.RawMessage `db:"custom_headers" json:"custom_headers"`
	CacheEnabled           bool            `db:"cache_enabled" json:"cache_enabled"`
	CacheTTL               int             `db:"cache_ttl" json:"cache_ttl"`
	CacheMethods           pq.StringArray  `db:"cache_methods" json:"cache_methods"`
	CacheIgnoreHeaders     bool            `db:"cache_ignore_headers" json:"cache_ignore_headers"`
	HSTSEnabled            bool            `db:"hsts_enabled" json:"hsts_enabled"`
	HSTSMaxAge             int             `db:"hsts_max_age" json:"hsts_max_age"`
	HSTSIncludeSubdomains bool            `db:"hsts_include_subdomains" json:"hsts_include_subdomains"`
	HSTSPreload            bool            `db:"hsts_preload" json:"hsts_preload"`
	BrotliEnabled          bool            `db:"brotli_enabled" json:"brotli_enabled"`
	HTTP3Enabled           bool            `db:"http3_enabled" json:"http3_enabled"`
	HideServerTokens       bool            `db:"hide_server_tokens" json:"hide_server_tokens"`
	SecurityHeadersEnabled bool            `db:"security_headers_enabled" json:"security_headers_enabled"`
	ClientBodyBufferSize   int             `db:"client_body_buffer_size" json:"client_body_buffer_size"`
	CreatedAt              time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt              time.Time       `db:"updated_at" json:"updated_at"`
}

// Fixed vhostDBResponse for any RawMessage issues: region_whitelist/blacklist should be handled carefully.
// I'll stick to the previous successful format.

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

// Handler functions

func (h *VHostHandler) ListVHosts(c *gin.Context) {
	var vhosts []vhostDBResponse
	query := `SELECT id::text, name, COALESCE(type, 'proxy') as type, domain, backend_url, backends::text as backends, COALESCE(load_balance_method, 'round_robin') as load_balance_method, custom_config, ssl_enabled, ssl_certificate_id::text, ssl_cert_path, ssl_key_path, enabled, websocket_enabled, http_version, tls_version, max_upload_size, client_body_buffer_size, proxy_read_timeout, proxy_connect_timeout, bot_detection_enabled, bot_detection_type, recaptcha_version, rate_limit_enabled, rate_limit_requests, rate_limit_window, region_filtering_enabled, COALESCE(region_whitelist, '{}') as region_whitelist, COALESCE(region_blacklist, '{}') as region_blacklist, COALESCE(defense_mode, 'defense') as defense_mode, custom_headers, cache_enabled, cache_ttl, COALESCE(cache_methods, '{}') AS cache_methods, cache_ignore_headers, hsts_enabled, hsts_max_age, hsts_include_subdomains, hsts_preload, brotli_enabled, http3_enabled, hide_server_tokens, security_headers_enabled, created_at, updated_at FROM vhosts ORDER BY created_at DESC`
	if err := h.db.Select(&vhosts, query); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	locationMap := h.fetchAllVHostLocationsFromDB()
	var response []map[string]interface{}
	for _, vhost := range vhosts {
		response = append(response, h.buildVHostResponse(vhost, locationMap[vhost.ID], nil))
	}
	c.JSON(http.StatusOK, response)
}

func (h *VHostHandler) GetVHost(c *gin.Context) {
	id := c.Param("id")
	var vhost vhostDBResponse
	query := `SELECT id::text, name, COALESCE(type, 'proxy') as type, domain, backend_url, backends::text as backends, COALESCE(load_balance_method, 'round_robin') as load_balance_method, custom_config, ssl_enabled, ssl_certificate_id::text, ssl_cert_path, ssl_key_path, enabled, websocket_enabled, http_version, tls_version, max_upload_size, client_body_buffer_size, proxy_read_timeout, proxy_connect_timeout, bot_detection_enabled, bot_detection_type, recaptcha_version, rate_limit_enabled, rate_limit_requests, rate_limit_window, region_filtering_enabled, COALESCE(region_whitelist, '{}') as region_whitelist, COALESCE(region_blacklist, '{}') as region_blacklist, COALESCE(defense_mode, 'defense') as defense_mode, custom_headers, cache_enabled, cache_ttl, COALESCE(cache_methods, '{}') AS cache_methods, cache_ignore_headers, hsts_enabled, hsts_max_age, hsts_include_subdomains, hsts_preload, brotli_enabled, http3_enabled, hide_server_tokens, security_headers_enabled, created_at, updated_at FROM vhosts WHERE id = $1`
	if err := h.db.Get(&vhost, query, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "VHost not found"})
		return
	}
	locations := h.fetchVHostLocations(id)
	ipGroupIDs := h.fetchVHostIPGroupIDs(id)
	c.JSON(http.StatusOK, h.buildVHostResponse(vhost, locations, ipGroupIDs))
}

func (h *VHostHandler) CreateVHost(c *gin.Context) {
	var input models.VHostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sanitizeVHostInput(&input)

	// Check if name or domain already exists
	var count int
	if err := h.db.Get(&count, "SELECT count(*) FROM vhosts WHERE name = $1 OR domain = $2", input.Name, input.Domain); err == nil && count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "VHost with the same name or domain already exists"})
		return
	}

	if err := h.validateSSLCert(input.SSLCertificateID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.db.Beginx()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction: " + err.Error()})
		return
	}
	defer tx.Rollback()

	hJSON, bJSON, _ := h.marshalVHostInputs(input)
	query := `INSERT INTO vhosts (id, name, type, domain, backend_url, backends, load_balance_method, custom_config, ssl_enabled, ssl_certificate_id, ssl_cert_path, ssl_key_path, enabled, websocket_enabled, http_version, tls_version, max_upload_size, client_body_buffer_size, proxy_read_timeout, proxy_connect_timeout, bot_detection_enabled, bot_detection_type, recaptcha_version, rate_limit_enabled, rate_limit_requests, rate_limit_window, region_whitelist, region_blacklist, region_filtering_enabled, defense_mode, custom_headers, cache_enabled, cache_ttl, cache_methods, cache_ignore_headers, hsts_enabled, hsts_max_age, hsts_include_subdomains, hsts_preload, brotli_enabled, http3_enabled, hide_server_tokens, security_headers_enabled, created_at, updated_at) VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41, $42, $43, $44) RETURNING id`
	var id string
	if err := tx.QueryRow(query, input.Name, input.Type, input.Domain, input.BackendURL, bJSON, input.LoadBalanceMethod, input.CustomConfig, input.SSLEnabled, input.SSLCertificateID, input.SSLCertPath, input.SSLKeyPath, input.Enabled, input.WebsocketEnabled, input.HTTPVersion, input.TLSVersion, input.MaxUploadSize, input.ClientBodyBufferSize, input.ProxyReadTimeout, input.ProxyConnectTimeout, input.BotDetectionEnabled, input.BotDetectionType, input.RecaptchaVersion, input.RateLimitEnabled, input.RateLimitRequests, input.RateLimitWindow, pq.Array(input.RegionWhitelist), pq.Array(input.RegionBlacklist), input.RegionFilteringEnabled, input.DefenseMode, hJSON, input.CacheEnabled, input.CacheTTL, pq.Array(input.CacheMethods), input.CacheIgnoreHeaders, input.HSTSEnabled, input.HSTSMaxAge, input.HSTSIncludeSubdomains, input.HSTSPreload, input.BrotliEnabled, input.HTTP3Enabled, input.HideServerTokens, input.SecurityHeadersEnabled, time.Now(), time.Now()).Scan(&id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert VHost: " + err.Error()})
		return
	}

	if err := h.upsertVHostAssociationsTx(tx, id, input, false); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save associations: " + err.Error()})
		return
	}

	// Generate Nginx config before committing
	// Build VHost object manually from input and ID to avoid cross-transaction fetch issues
	v := h.buildPreviewVHost(input).VHost
	v.ID = id
	if err := h.nginxConfigService.GenerateVHostConfig(v); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Nginx config: " + err.Error()})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
		return
	}

	h.finalizeVHostChange(id, input)
	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "VHost created successfully"})
}

func (h *VHostHandler) UpdateVHost(c *gin.Context) {
	id := c.Param("id")
	var input models.VHostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sanitizeVHostInput(&input)

	// Check if name or domain already exists in OTHER vhosts
	var count int
	if err := h.db.Get(&count, "SELECT count(*) FROM vhosts WHERE (name = $1 OR domain = $2) AND id != $3", input.Name, input.Domain, id); err == nil && count > 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Another VHost with the same name or domain already exists"})
		return
	}

	if err := h.validateSSLCert(input.SSLCertificateID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tx, err := h.db.Beginx()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to start transaction: " + err.Error()})
		return
	}
	defer tx.Rollback()

	hJSON, bJSON, _ := h.marshalVHostInputs(input)
	query := `UPDATE vhosts SET name = $1, type = $2, domain = $3, backend_url = $4, backends = $5, load_balance_method = $6, custom_config = $7, ssl_enabled = $8, ssl_certificate_id = $9, ssl_cert_path = $10, ssl_key_path = $11, enabled = $12, websocket_enabled = $13, http_version = $14, tls_version = $15, max_upload_size = $16, client_body_buffer_size = $17, proxy_read_timeout = $18, proxy_connect_timeout = $19, bot_detection_enabled = $20, bot_detection_type = $21, recaptcha_version = $22, rate_limit_enabled = $23, rate_limit_requests = $24, rate_limit_window = $25, region_whitelist = $26, region_blacklist = $27, region_filtering_enabled = $28, defense_mode = $29, custom_headers = $30, cache_enabled = $31, cache_ttl = $32, cache_methods = $33, cache_ignore_headers = $34, hsts_enabled = $35, hsts_max_age = $36, hsts_include_subdomains = $37, hsts_preload = $38, brotli_enabled = $39, http3_enabled = $40, hide_server_tokens = $41, security_headers_enabled = $42, updated_at = $43 WHERE id = $44`
	if _, err := tx.Exec(query, input.Name, input.Type, input.Domain, input.BackendURL, bJSON, input.LoadBalanceMethod, input.CustomConfig, input.SSLEnabled, input.SSLCertificateID, input.SSLCertPath, input.SSLKeyPath, input.Enabled, input.WebsocketEnabled, input.HTTPVersion, input.TLSVersion, input.MaxUploadSize, input.ClientBodyBufferSize, input.ProxyReadTimeout, input.ProxyConnectTimeout, input.BotDetectionEnabled, input.BotDetectionType, input.RecaptchaVersion, input.RateLimitEnabled, input.RateLimitRequests, input.RateLimitWindow, pq.Array(input.RegionWhitelist), pq.Array(input.RegionBlacklist), input.RegionFilteringEnabled, input.DefenseMode, hJSON, input.CacheEnabled, input.CacheTTL, pq.Array(input.CacheMethods), input.CacheIgnoreHeaders, input.HSTSEnabled, input.HSTSMaxAge, input.HSTSIncludeSubdomains, input.HSTSPreload, input.BrotliEnabled, input.HTTP3Enabled, input.HideServerTokens, input.SecurityHeadersEnabled, time.Now(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update VHost: " + err.Error()})
		return
	}

	if err := h.upsertVHostAssociationsTx(tx, id, input, true); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save associations: " + err.Error()})
		return
	}

	// Generate Nginx config before committing
	// Build VHost object manually from input and ID to avoid cross-transaction fetch issues
	v := h.buildPreviewVHost(input).VHost
	v.ID = id

	if err := h.nginxConfigService.GenerateVHostConfig(v); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Nginx config: " + err.Error()})
		return
	}

	if err := tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit transaction: " + err.Error()})
		return
	}

	h.finalizeVHostChange(id, input)
	c.JSON(http.StatusOK, gin.H{"message": "VHost updated successfully"})
}

func (h *VHostHandler) DeleteVHost(c *gin.Context) {
	id := c.Param("id")
	vhost, err := h.vhostService.GetVHostByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": constants.ErrVHostNotFound})
		return
	}
	if _, err := h.db.Exec("DELETE FROM vhosts WHERE id = $1", id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vhost: " + err.Error()})
		return
	}
	_ = h.nginxConfigService.DeleteVHostConfig(vhost.Domain)
	h.finalizeVHostChange("", models.VHostInput{})
	c.JSON(http.StatusOK, gin.H{"message": "VHost deleted successfully"})
}

// Helpers

func (h *VHostHandler) fetchAllVHostLocationsFromDB() map[string][]map[string]interface{} {
	type Loc struct {
		VHostID           string  `db:"vhost_id"`
		Path              string  `db:"path"`
		ProxyPass         *string `db:"proxy_pass"`
		CustomConfig      *string `db:"custom_config"`
		WebSocketEnabled  bool    `db:"websocket_enabled"`
		Backends          *string `db:"backends"`
		LoadBalanceMethod *string `db:"load_balance_method"`
	}
	var all []Loc
	_ = h.db.Select(&all, "SELECT vhost_id::text, path, proxy_pass, custom_config, COALESCE(websocket_enabled, false) as websocket_enabled, backends::text as backends, COALESCE(load_balance_method, 'round_robin') as load_balance_method FROM vhost_locations WHERE enabled = true ORDER BY created_at ASC")
	m := make(map[string][]map[string]interface{})
	for _, l := range all {
		var b []string
		if l.Backends != nil && *l.Backends != "" && *l.Backends != "[]" {
			_ = json.Unmarshal([]byte(*l.Backends), &b)
		}
		m[l.VHostID] = append(m[l.VHostID], map[string]interface{}{
			"path": l.Path, "proxy_pass": l.ProxyPass, "config": l.CustomConfig, "websocket_enabled": l.WebSocketEnabled, "backends": b, "load_balance_method": l.LoadBalanceMethod,
		})
	}
	return m
}

func (h *VHostHandler) fetchVHostLocations(vhostID string) []map[string]interface{} {
	type Loc struct {
		Path string `db:"path"`; ProxyPass *string `db:"proxy_pass"`; CustomConfig *string `db:"custom_config"`; WebSocketEnabled bool `db:"websocket_enabled"`; Backends *string `db:"backends"`; LoadBalanceMethod *string `db:"load_balance_method"`
	}
	var locs []Loc
	_ = h.db.Select(&locs, "SELECT path, proxy_pass, custom_config, COALESCE(websocket_enabled, false) as websocket_enabled, backends::text as backends, COALESCE(load_balance_method, 'round_robin') as load_balance_method FROM vhost_locations WHERE vhost_id = $1 AND enabled = true ORDER BY length(path) DESC", vhostID)
	res := []map[string]interface{}{}
	for _, l := range locs {
		var b []string
		if l.Backends != nil && *l.Backends != "" && *l.Backends != "[]" {
			_ = json.Unmarshal([]byte(*l.Backends), &b)
		}
		res = append(res, map[string]interface{}{"path": l.Path, "proxy_pass": l.ProxyPass, "config": l.CustomConfig, "websocket_enabled": l.WebSocketEnabled, "backends": b, "load_balance_method": l.LoadBalanceMethod})
	}
	return res
}

func (h *VHostHandler) fetchVHostIPGroupIDs(vhostID string) []string {
	var ids []string
	_ = h.db.Select(&ids, "SELECT ip_group_id::text FROM ip_group_vhosts WHERE vhost_id = $1", vhostID)
	if ids == nil { return []string{} }
	return ids
}

func (h *VHostHandler) buildVHostResponse(vhost vhostDBResponse, locs []map[string]interface{}, ipIDs []string) map[string]interface{} {
	res := map[string]interface{}{
		"id": vhost.ID, "name": vhost.Name, "type": vhost.Type, "domain": vhost.Domain, "backend_url": vhost.BackendURL, "backends": parseJSONStringArray(vhost.Backends),
		"load_balance_method": vhost.LoadBalanceMethod, "custom_config": vhost.CustomConfig, "ssl_enabled": vhost.SSLEnabled, "ssl_certificate_id": vhost.SSLCertificateID, "ssl_cert_path": vhost.SSLCertPath, "ssl_key_path": vhost.SSLKeyPath, "enabled": vhost.Enabled,
		"websocket_enabled": vhost.WebsocketEnabled, "http_version": vhost.HTTPVersion, "tls_version": vhost.TLSVersion, "max_upload_size": vhost.MaxUploadSize, "proxy_read_timeout": vhost.ProxyReadTimeout, "proxy_connect_timeout": vhost.ProxyConnectTimeout,
		"bot_detection_enabled": vhost.BotDetectionEnabled, "bot_detection_type": vhost.BotDetectionType, "recaptcha_version": vhost.RecaptchaVersion, "rate_limit_enabled": vhost.RateLimitEnabled, "rate_limit_requests": vhost.RateLimitRequests, "rate_limit_window": vhost.RateLimitWindow,
		"region_filtering_enabled": vhost.RegionFilteringEnabled, "region_whitelist": []string(vhost.RegionWhitelist), "region_blacklist": []string(vhost.RegionBlacklist), "defense_mode": vhost.DefenseMode, "custom_headers": vhost.CustomHeaders,
		"cache_enabled": vhost.CacheEnabled, "cache_ttl": vhost.CacheTTL, "cache_methods": []string(vhost.CacheMethods), "cache_ignore_headers": vhost.CacheIgnoreHeaders,
		"hsts_enabled": vhost.HSTSEnabled, "hsts_max_age": vhost.HSTSMaxAge, "hsts_include_subdomains": vhost.HSTSIncludeSubdomains, "hsts_preload": vhost.HSTSPreload,
		"brotli_enabled": vhost.BrotliEnabled, "http3_enabled": vhost.HTTP3Enabled, "hide_server_tokens": vhost.HideServerTokens, "security_headers_enabled": vhost.SecurityHeadersEnabled,
		"client_body_buffer_size": vhost.ClientBodyBufferSize,
		"created_at": vhost.CreatedAt, "updated_at": vhost.UpdatedAt, "custom_locations": locs,
	}
	if ipIDs != nil { res["ip_group_ids"] = ipIDs }
	return res
}

func (h *VHostHandler) validateSSLCert(id *string) error {
	if id != nil && *id != "" {
		if _, err := h.certService.GetCertificate(*id); err != nil { return fmt.Errorf("invalid SSL certificate ID") }
	}
	return nil
}

func (h *VHostHandler) marshalVHostInputs(input models.VHostInput) ([]byte, []byte, error) {
	headers, _ := json.Marshal(input.CustomHeaders)
	backends, _ := json.Marshal(input.Backends)
	return headers, backends, nil
}

func (h *VHostHandler) upsertVHostAssociationsTx(execer sqlx.Execer, id string, input models.VHostInput, update bool) error {
	if update {
		_, err := execer.Exec("DELETE FROM ip_group_vhosts WHERE vhost_id = $1", id)
		if err != nil { return err }
		_, err = execer.Exec("DELETE FROM vhost_locations WHERE vhost_id = $1", id)
		if err != nil { return err }
	}
	for _, pid := range input.IPGroupIDs {
		if pid != "" {
			_, err := execer.Exec("INSERT INTO ip_group_vhosts (ip_group_id, vhost_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", pid, id)
			if err != nil { return err }
		}
	}
	for _, loc := range input.CustomLocations {
		path, _ := loc["path"].(string); pPass, _ := loc["proxy_pass"]
		ws, _ := loc["websocket_enabled"].(bool); var bj []byte = []byte("[]")
		if b, ok := loc["backends"].([]interface{}); ok { bj, _ = json.Marshal(b) }
		m, _ := loc["load_balance_method"].(string); if m == "" { m = "round_robin" }
		_, err := execer.Exec(`INSERT INTO vhost_locations (id, vhost_id, path, backend_url, proxy_pass, custom_config, websocket_enabled, backends, load_balance_method, enabled, created_at, updated_at) VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, true, $9, $10)`, id, path, pPass, pPass, loc["config"], ws, bj, m, time.Now(), time.Now())
		if err != nil { return err }
	}
	return nil
}

func (h *VHostHandler) finalizeVHostChange(id string, input models.VHostInput) {
	if id != "" {
		if input.SSLEnabled && input.SSLCertificateID != nil && *input.SSLCertificateID != "" {
			if c, err := h.certService.GetCertificate(*input.SSLCertificateID); err == nil {
				_ = h.certService.SaveCertificateFiles(*input.SSLCertificateID, []byte(c.CertContent), []byte(c.KeyContent))
			}
		}
		if v, err := h.vhostService.GetVHostByID(id); err == nil {
			_ = h.nginxConfigService.GenerateVHostConfig(v)
		}
	}
	h.reloadNginx()
	if h.proxyReloader != nil { _ = h.proxyReloader.ReloadVHosts() }
}

func (h *VHostHandler) reloadNginx() {
	_ = os.WriteFile("/data/nginx/.reload", []byte(time.Now().Format(time.RFC3339)), 0644)
}

// Handler specific endpoints

func (h *VHostHandler) PreviewVHostConfig(c *gin.Context) {
	var input models.VHostInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	sanitizeVHostInput(&input)
	v := h.buildPreviewVHost(input)
	content, err := h.nginxConfigService.GenerateVHostConfigContent(v)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"config": content})
}

func (h *VHostHandler) buildPreviewVHost(input models.VHostInput) *services.VHostWithLocations {
	sID := ""
	if input.SSLCertificateID != nil { sID = *input.SSLCertificateID }
	ch, _ := json.Marshal(input.CustomHeaders)

	v := &services.VHostWithLocations{
		VHost: &models.VHost{
			Name: input.Name, Type: input.Type, Domain: input.Domain, BackendURL: input.BackendURL, Backends: input.Backends, LoadBalanceMethod: input.LoadBalanceMethod, CustomConfig: input.CustomConfig,
			SSLEnabled: input.SSLEnabled, SSLCertificateID: sID, SSLCertPath: input.SSLCertPath, SSLKeyPath: input.SSLKeyPath,
			Enabled: input.Enabled, WebsocketEnabled: input.WebsocketEnabled, HTTPVersion: input.HTTPVersion, TLSVersion: input.TLSVersion,
			MaxUploadSize: input.MaxUploadSize, ProxyReadTimeout: input.ProxyReadTimeout, ProxyConnectTimeout: input.ProxyConnectTimeout,
			BotDetectionEnabled: input.BotDetectionEnabled, BotDetectionType: input.BotDetectionType, RecaptchaVersion: input.RecaptchaVersion,
			RateLimitEnabled: input.RateLimitEnabled, RateLimitRequests: input.RateLimitRequests, RateLimitWindow: input.RateLimitWindow,
			RegionWhitelist: input.RegionWhitelist, RegionBlacklist: input.RegionBlacklist, RegionFilteringEnabled: input.RegionFilteringEnabled,
			DefenseMode: input.DefenseMode, CustomHeaders: ch,
			CacheEnabled: input.CacheEnabled, CacheTTL: input.CacheTTL, CacheMethods: input.CacheMethods, CacheIgnoreHeaders: input.CacheIgnoreHeaders,
			HSTSEnabled: input.HSTSEnabled, HSTSMaxAge: input.HSTSMaxAge, HSTSIncludeSubdomains: input.HSTSIncludeSubdomains, HSTSPreload: input.HSTSPreload,
			BrotliEnabled: input.BrotliEnabled, HTTP3Enabled: input.HTTP3Enabled, HideServerTokens: input.HideServerTokens, SecurityHeadersEnabled: input.SecurityHeadersEnabled,
			ClientBodyBufferSize:   input.ClientBodyBufferSize,
		},
		CustomLocations: []services.CustomLocation{},
	}
	re := regexp.MustCompile(`[^a-zA-Z0-9]`)
	uName := re.ReplaceAllString(input.Domain, "_")
	v.UpstreamName = uName
	ab := input.Backends
	if len(ab) > 0 && input.BackendURL != "" { ab = append([]string{input.BackendURL}, ab...) }
	if v.VHost.Type == "proxy" && len(ab) > 0 { v.HasUpstream = true; v.Backends = ab }
	v.CustomLocations = h.processPreviewLocations(input.CustomLocations, uName, re)
	return v
}

func (h *VHostHandler) processPreviewLocations(locs []map[string]interface{}, uName string, re *regexp.Regexp) []services.CustomLocation {
	var res []services.CustomLocation
	for _, l := range locs {
		p, _ := l["path"].(string); pp, _ := l["proxy_pass"].(string); cfg, _ := l["config"].(string); ws, _ := l["websocket_enabled"].(bool)
		cl := services.CustomLocation{Path: p, ProxyPass: pp, CustomConfig: cfg, WebSocketEnabled: ws}
		if bs, ok := l["backends"].([]interface{}); ok {
			var lbs []string
			for _, bi := range bs { if s, ok := bi.(string); ok { lbs = append(lbs, s) } }
			if len(lbs) > 0 { cl.Backends = lbs; cl.HasUpstream = true; cl.UpstreamName = uName + "_loc_" + re.ReplaceAllString(p, "_") }
		}
		res = append(res, cl)
	}
	return res
}

func (h *VHostHandler) GetVHostConfig(c *gin.Context) {
	domain := c.Param("domain")
	configPath := fmt.Sprintf("/data/nginx/config/%s.conf", domain)
	content, err := os.ReadFile(configPath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Config file not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"domain": domain, "content": string(content), "path": configPath})
}

func (h *VHostHandler) UpdateVHostConfig(c *gin.Context) {
	domain := c.Param("domain")
	var req struct { Content string `json:"content" binding:"required"` }
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	configPath := fmt.Sprintf("/data/nginx/config/%s.conf", domain)
	err := os.WriteFile(configPath, []byte(req.Content), 0644)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	h.reloadNginx()
	c.JSON(http.StatusOK, gin.H{"message": "Config updated successfully", "domain": domain})
}

func (h *VHostHandler) RegenerateAllConfigs(c *gin.Context) {
	vhosts, err := h.vhostService.ListVHosts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch vhosts: " + err.Error()})
		return
	}
	var regenerated []string
	for _, vhost := range vhosts {
		if vhost.Enabled {
			if err := h.nginxConfigService.GenerateVHostConfig(vhost); err == nil {
				regenerated = append(regenerated, vhost.Domain)
			}
		}
	}
	h.reloadNginx()
	if h.proxyReloader != nil { _ = h.proxyReloader.ReloadVHosts() }
	c.JSON(http.StatusOK, gin.H{"message": "Configs regenerated successfully", "regenerated": regenerated, "count": len(regenerated)})
}
