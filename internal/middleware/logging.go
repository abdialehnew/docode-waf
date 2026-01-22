package middleware

import (
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/oschwald/geoip2-golang"
)

var (
	geoipDB     *geoip2.Reader
	geoipDBOnce sync.Once
)

// initGeoIP initializes the GeoIP database
func initGeoIP() {
	geoipDBOnce.Do(func() {
		// Try multiple paths for GeoIP database
		paths := []string{
			"/GeoLite2-Country.mmdb",  // Docker container path
			"GeoLite2-Country.mmdb",   // Local development path
			"./GeoLite2-Country.mmdb", // Current directory
		}

		for _, path := range paths {
			db, err := geoip2.Open(path)
			if err == nil {
				geoipDB = db
				log.Printf("GeoIP database loaded successfully from: %s", path)
				return
			}
		}
		log.Printf("Warning: Failed to load GeoIP database from any path")
	})
}

// getCountryCode extracts country code from IP address using GeoIP2
// Returns country code or "XX" if unknown
func getCountryCode(ip string) string {
	// Ensure GeoIP database is initialized
	initGeoIP()

	// Parse IP
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return "XX"
	}

	// Check if private/local IP
	if parsedIP.IsPrivate() || parsedIP.IsLoopback() {
		return "XX"
	}

	// Lookup country from GeoIP database
	if geoipDB != nil {
		record, err := geoipDB.Country(parsedIP)
		if err == nil && record.Country.IsoCode != "" {
			return record.Country.IsoCode
		}
	}

	// Return XX if lookup fails
	return "XX"
}

// LoggingMiddleware logs all HTTP traffic
func LoggingMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate response time
		duration := time.Since(start)

		// Log to database asynchronously
		go logTraffic(db, c, duration)
	}
}

// detectAttackType analyzes request for attack patterns
// Skips detection for private/local IPs to avoid false positives during testing
func detectAttackType(c *gin.Context) (bool, string) {
	// Skip attack detection for private/local IPs
	clientIP := net.ParseIP(c.ClientIP())
	if clientIP != nil && (clientIP.IsPrivate() || clientIP.IsLoopback()) {
		return false, ""
	}

	url := c.Request.URL.String()
	userAgent := c.GetHeader("User-Agent")

	// SQL Injection patterns
	sqlPatterns := []string{"' OR '1'='1", "' OR 1=1", "UNION SELECT", "'; DROP TABLE",
		"admin'--", "' OR ''='", "1' AND '1'='1", "SELECT * FROM"}
	for _, pattern := range sqlPatterns {
		if strings.Contains(strings.ToUpper(url), strings.ToUpper(pattern)) {
			return true, "SQL Injection"
		}
	}

	// XSS patterns
	xssPatterns := []string{"<script>", "</script>", "javascript:", "onerror=", "onload=",
		"<img", "alert(", "<iframe"}
	for _, pattern := range xssPatterns {
		if strings.Contains(strings.ToLower(url), strings.ToLower(pattern)) {
			return true, "XSS"
		}
	}

	// Path Traversal
	pathTraversalPatterns := []string{"../", "..\\", "/etc/passwd", "windows/system32", "../../"}
	for _, pattern := range pathTraversalPatterns {
		if strings.Contains(strings.ToLower(url), strings.ToLower(pattern)) {
			return true, "Path Traversal"
		}
	}

	// Command Injection
	cmdPatterns := []string{";ls", ";cat", ";whoami", "|ls", "|cat", "&ls", "$("}
	for _, pattern := range cmdPatterns {
		if strings.Contains(url, pattern) {
			return true, "Command Injection"
		}
	}

	// Admin Panel Scanning - use more specific patterns to avoid false positives
	// Check for actual admin paths, not just keywords in source code files
	urlLower := strings.ToLower(url)

	// Exact path matches (start of path)
	adminPaths := []string{"/admin", "/administrator", "/wp-admin", "/phpmyadmin",
		"/cpanel", "/admin.php", "/adminpanel", "/backend", "/management"}
	for _, path := range adminPaths {
		// Check if URL starts with admin path or has it after domain
		if strings.HasPrefix(urlLower, path) || strings.Contains(urlLower, "://"+c.Request.Host+path) {
			return true, "Admin Scan"
		}
	}

	// Common admin file patterns (avoid matching .tsx, .ts, .jsx, .js source files)
	adminFilePatterns := []string{"/admin/login", "/admin/index", "/login.php", "/admin.asp"}
	for _, pattern := range adminFilePatterns {
		if strings.Contains(urlLower, pattern) {
			return true, "Admin Scan"
		}
	}

	// Exclude source code files from detection
	sourceFileExtensions := []string{".tsx", ".ts", ".jsx", ".js", ".vue", ".py", ".go", ".java"}
	for _, ext := range sourceFileExtensions {
		if strings.HasSuffix(urlLower, ext) {
			// Skip attack detection for source code files
			break
		}
	}

	// Bot Detection
	botPatterns := []string{"bot", "crawler", "spider", "python", "curl", "wget"}
	for _, pattern := range botPatterns {
		if strings.Contains(strings.ToLower(userAgent), pattern) {
			return true, "Bot Traffic"
		}
	}

	return false, ""
}

func logTraffic(db *sqlx.DB, c *gin.Context, duration time.Duration) {
	// Detect attack
	isAttack, attackType := detectAttackType(c)
	blocked := c.GetBool("blocked") || c.Writer.Status() == 403

	query := `
		INSERT INTO traffic_logs (
			id, timestamp, client_ip, method, url, status_code, 
			response_time, bytes_sent, user_agent, blocked, block_reason,
			is_attack, attack_type, country_code, host
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
		)
	`

	blockReason := ""
	if blocked {
		if val, exists := c.Get("block_reason"); exists {
			blockReason = val.(string)
		}
	}

	countryCode := getCountryCode(c.ClientIP())

	// Get the HTTP Host header and try to find the matching vhost domain
	httpHost := c.Request.Host
	host := lookupVHostDomain(db, httpHost)

	_, err := db.Exec(query,
		time.Now(),
		c.ClientIP(),
		c.Request.Method,
		c.Request.URL.String(), // Changed from Path to String to include query params
		c.Writer.Status(),
		int(duration.Milliseconds()),
		c.Writer.Size(),
		c.GetHeader("User-Agent"),
		blocked,
		blockReason,
		isAttack,
		attackType,
		countryCode,
		host,
	)

	if err != nil {
		// Log error but don't fail the request
		println("Failed to log traffic:", err.Error())
	}
}

// lookupVHostDomain finds the vhost domain from database based on HTTP Host header
// If Host is an IP or doesn't match any vhost, returns the original Host value (for fallback)
// If a matching vhost is found by domain, returns the domain name
func lookupVHostDomain(db *sqlx.DB, httpHost string) string {
	// Remove port if present (host:port)
	hostOnly := httpHost
	if idx := strings.LastIndex(httpHost, ":"); idx != -1 {
		hostOnly = httpHost[:idx]
	}

	// First, check if the host is already a domain in our vhosts table
	var domain string
	err := db.Get(&domain, "SELECT domain FROM vhosts WHERE domain = $1 AND enabled = true LIMIT 1", hostOnly)
	if err == nil {
		return domain
	}

	// If not found by exact domain match, try to find any enabled vhost
	// This handles cases where request comes via IP address
	// We'll return the first enabled vhost domain as it's likely the primary one
	err = db.Get(&domain, "SELECT domain FROM vhosts WHERE enabled = true ORDER BY created_at ASC LIMIT 1")
	if err == nil {
		return domain
	}

	// Fallback to original host if no vhost found
	return httpHost
}
