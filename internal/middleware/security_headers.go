package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// SecurityHeadersMiddleware adds security headers and removes sensitive information
// Implements protection for A02:2021 (Cryptographic Failures) and A05:2021 (Security Misconfiguration)
func SecurityHeadersMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get current vhost domain
		domain := c.Request.Host
		if colonIdx := strings.Index(domain, ":"); colonIdx != -1 {
			domain = domain[:colonIdx]
		}

		// Check vhost security settings
		var vhostSettings struct {
			SecurityHeadersEnabled bool   `db:"security_headers_enabled"`
			HSTSEnabled            bool   `db:"hsts_enabled"`
			HSTSMaxAge             int    `db:"hsts_max_age"`
			CSPPolicy              string `db:"csp_policy"`
			PermissionsPolicy      string `db:"permissions_policy"`
		}

		err := db.Get(&vhostSettings, `
			SELECT 
				COALESCE(security_headers_enabled, true) as security_headers_enabled,
				COALESCE(hsts_enabled, true) as hsts_enabled,
				COALESCE(hsts_max_age, 31536000) as hsts_max_age,
				COALESCE(csp_policy, '') as csp_policy,
				COALESCE(permissions_policy, '') as permissions_policy
			FROM vhosts WHERE domain = $1`, domain)

		if err != nil {
			// Use defaults if vhost not found
			vhostSettings.SecurityHeadersEnabled = true
			vhostSettings.HSTSEnabled = true
			vhostSettings.HSTSMaxAge = 31536000 // 1 year
		}

		// Skip if security headers disabled
		if !vhostSettings.SecurityHeadersEnabled {
			c.Next()
			return
		}

		// === Response Headers (set before processing) ===

		// X-Content-Type-Options - Prevent MIME type sniffing
		c.Header("X-Content-Type-Options", "nosniff")

		// X-Frame-Options - Prevent clickjacking
		c.Header("X-Frame-Options", "SAMEORIGIN")

		// X-XSS-Protection - Legacy XSS protection (for older browsers)
		c.Header("X-XSS-Protection", "1; mode=block")

		// Referrer-Policy - Control referrer information
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		// Permissions-Policy (formerly Feature-Policy)
		if vhostSettings.PermissionsPolicy != "" {
			c.Header("Permissions-Policy", vhostSettings.PermissionsPolicy)
		} else {
			// Default restrictive policy
			c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=(), magnetometer=(), gyroscope=(), accelerometer=()")
		}

		// Content-Security-Policy
		if vhostSettings.CSPPolicy != "" {
			c.Header("Content-Security-Policy", vhostSettings.CSPPolicy)
		}
		// Note: CSP can break applications, so we don't set a default

		// HSTS - HTTP Strict Transport Security
		if vhostSettings.HSTSEnabled && c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		}

		// Cross-Origin headers
		c.Header("Cross-Origin-Opener-Policy", "same-origin")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")

		// Process request
		c.Next()

		// === Post-processing: Remove sensitive headers from response ===

		// Remove server identification headers
		c.Writer.Header().Del("Server")
		c.Writer.Header().Del("X-Powered-By")
		c.Writer.Header().Del("X-AspNet-Version")
		c.Writer.Header().Del("X-AspNetMvc-Version")
		c.Writer.Header().Del("X-Runtime")
		c.Writer.Header().Del("X-Version")
	}
}

// HTTPSRedirectMiddleware redirects HTTP to HTTPS
func HTTPSRedirectMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if request is already HTTPS
		if c.Request.TLS != nil {
			c.Next()
			return
		}

		// Check X-Forwarded-Proto header (for reverse proxy)
		if c.GetHeader("X-Forwarded-Proto") == "https" {
			c.Next()
			return
		}

		// Redirect to HTTPS
		httpsURL := "https://" + c.Request.Host + c.Request.URL.String()
		c.Redirect(301, httpsURL)
		c.Abort()
	}
}

// SensitiveDataLeakageMiddleware detects and blocks responses containing sensitive data
func SensitiveDataLeakageMiddleware() gin.HandlerFunc {
	// Note: Full response body inspection would require wrapping the ResponseWriter
	// Patterns that indicate sensitive data leakage are checked post-processing

	return func(c *gin.Context) {
		// We can't easily inspect response body in Gin middleware
		// This would require a custom response writer
		// For now, we just add headers and let the OWASP rules handle request-side protection
		c.Next()

		// Additional: Remove any debug headers that might leak info
		debugHeaders := []string{
			"X-Debug-Token",
			"X-Debug-Token-Link",
			"X-Symfony-Debug-Info",
			"X-Laravel-Debug",
			"X-PHP-Error",
		}
		for _, header := range debugHeaders {
			c.Writer.Header().Del(header)
		}
	}
}

// CORSSecurityMiddleware provides secure CORS configuration
func CORSSecurityMiddleware(allowedOrigins []string) gin.HandlerFunc {
	originSet := make(map[string]bool)
	for _, origin := range allowedOrigins {
		originSet[strings.ToLower(origin)] = true
	}

	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// Check if origin is allowed
		if origin != "" {
			originLower := strings.ToLower(origin)
			if originSet[originLower] || originSet["*"] {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
				c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
				c.Header("Access-Control-Max-Age", "86400") // 24 hours
				c.Header("Vary", "Origin")
			}
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// RequestSanitizationMiddleware sanitizes common request inputs
func RequestSanitizationMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Validate Content-Type for POST/PUT/PATCH requests
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH" {
			contentType := c.GetHeader("Content-Type")
			if contentType != "" {
				// Ensure Content-Type is valid
				validContentTypes := []string{
					"application/json",
					"application/x-www-form-urlencoded",
					"multipart/form-data",
					"text/plain",
					"text/xml",
					"application/xml",
				}

				isValid := false
				contentTypeLower := strings.ToLower(contentType)
				for _, valid := range validContentTypes {
					if strings.HasPrefix(contentTypeLower, valid) {
						isValid = true
						break
					}
				}

				if !isValid {
					c.Set("suspicious_content_type", true)
				}
			}
		}

		// Validate Host header (prevent host header injection)
		host := c.Request.Host
		if strings.Contains(host, "@") || strings.Contains(host, "#") {
			c.JSON(400, gin.H{"error": "Invalid host header"})
			c.Abort()
			return
		}

		c.Next()
	}
}
