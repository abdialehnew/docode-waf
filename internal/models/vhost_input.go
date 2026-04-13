package models

type VHostInput struct {
	Name                   string                   `json:"name" binding:"required"`
	Type                   string                   `json:"type"`
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
	DefenseMode            string                   `json:"defense_mode"`
	CustomHeaders          map[string]interface{}   `json:"custom_headers"`
	CustomLocations        []map[string]interface{} `json:"custom_locations"`
	IPGroupIDs             []string                 `json:"ip_group_ids"`
	CacheEnabled           bool                     `json:"cache_enabled"`
	CacheTTL               int                      `json:"cache_ttl"`
	CacheMethods           []string                 `json:"cache_methods"`
	CacheIgnoreHeaders     bool                     `json:"cache_ignore_headers"`
	HSTSEnabled            bool                     `json:"hsts_enabled"`
	HSTSMaxAge             int                      `json:"hsts_max_age"`
	HSTSIncludeSubdomains bool                     `json:"hsts_include_subdomains"`
	HSTSPreload            bool                     `json:"hsts_preload"`
	BrotliEnabled          bool                     `json:"brotli_enabled"`
	HTTP3Enabled           bool                     `json:"http3_enabled"`
	HideServerTokens       bool                     `json:"hide_server_tokens"`
	SecurityHeadersEnabled bool                     `json:"security_headers_enabled"`
	ClientBodyBufferSize   int                      `json:"client_body_buffer_size"`
}
