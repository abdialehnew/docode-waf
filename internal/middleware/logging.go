package middleware

import (
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/aleh/docode-waf/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/oschwald/geoip2-golang"
)

var (
	geoipDB     *geoip2.Reader
	geoipDBOnce sync.Once
)

const (
	attackTypeAdminScan = "Admin Scan"
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
func LoggingMiddleware(db *sqlx.DB, banService *services.DynamicBanService, notificationService *services.NotificationService) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Process request
		c.Next()

		// Calculate response time
		duration := time.Since(start)

		// Log to database asynchronously
		go logTraffic(db, banService, notificationService, c, duration)
	}
}

// detectAttackType analyzes request for attack patterns
// Skips detection for private/local IPs to avoid false positives during testing
func detectAttackType(c *gin.Context) (bool, string) {
	// Skip attack detection for private/local IPs
	clientIP := net.ParseIP(GetRealClientIP(c))
	if clientIP != nil && (clientIP.IsPrivate() || clientIP.IsLoopback()) {
		return false, ""
	}

	url := c.Request.URL.String()
	userAgent := c.GetHeader("User-Agent")

	if match, _ := checkSQLInjection(url); match {
		return true, "SQL Injection"
	}
	if match, _ := checkXSS(url); match {
		return true, "XSS"
	}
	if match, _ := checkPathTraversal(url); match {
		return true, "Path Traversal"
	}
	if match, _ := checkCommandInjection(url); match {
		return true, "Command Injection"
	}
	if match, _ := checkAdminScan(url, c.Request.Host); match {
		return true, attackTypeAdminScan
	}
	if match, _ := checkBot(userAgent); match {
		return true, "Bot Traffic"
	}

	return false, ""
}

func checkSQLInjection(url string) (bool, string) {
	sqlPatterns := []string{"' OR '1'='1", "' OR 1=1", "UNION SELECT", "'; DROP TABLE",
		"admin'--", "' OR ''='", "1' AND '1'='1", "SELECT * FROM"}
	for _, pattern := range sqlPatterns {
		if strings.Contains(strings.ToUpper(url), strings.ToUpper(pattern)) {
			return true, "SQL Injection"
		}
	}
	return false, ""
}

func checkXSS(url string) (bool, string) {
	xssPatterns := []string{"<script>", "</script>", "javascript:", "onerror=", "onload=",
		"<img", "alert(", "<iframe"}
	for _, pattern := range xssPatterns {
		if strings.Contains(strings.ToLower(url), strings.ToLower(pattern)) {
			return true, "XSS"
		}
	}
	return false, ""
}

func checkPathTraversal(url string) (bool, string) {
	pathTraversalPatterns := []string{"../", "..\\", "/etc/passwd", "windows/system32", "../../"}
	for _, pattern := range pathTraversalPatterns {
		if strings.Contains(strings.ToLower(url), strings.ToLower(pattern)) {
			return true, "Path Traversal"
		}
	}
	return false, ""
}

func checkCommandInjection(url string) (bool, string) {
	cmdPatterns := []string{";ls", ";cat", ";whoami", "|ls", "|cat", "&ls", "$("}
	for _, pattern := range cmdPatterns {
		if strings.Contains(url, pattern) {
			return true, "Command Injection"
		}
	}
	return false, ""
}

func checkAdminScan(url, host string) (bool, string) {
	urlLower := strings.ToLower(url)

	// Exclude source code files from detection
	sourceFileExtensions := []string{".tsx", ".ts", ".jsx", ".js", ".vue", ".py", ".go", ".java"}
	for _, ext := range sourceFileExtensions {
		if strings.HasSuffix(urlLower, ext) {
			return false, ""
		}
	}

	// Exact path matches (start of path)
	adminPaths := []string{"/admin", "/administrator", "/wp-admin", "/phpmyadmin",
		"/cpanel", "/admin.php", "/adminpanel", "/backend", "/management"}
	for _, path := range adminPaths {
		if strings.HasPrefix(urlLower, path) || strings.Contains(urlLower, "://"+host+path) {
			return true, attackTypeAdminScan
		}
	}

	// Common admin file patterns
	adminFilePatterns := []string{"/admin/login", "/admin/index", "/login.php", "/admin.asp"}
	for _, pattern := range adminFilePatterns {
		if strings.Contains(urlLower, pattern) {
			return true, attackTypeAdminScan
		}
	}
	return false, ""
}

func checkBot(userAgent string) (bool, string) {
	botPatterns := []string{"bot", "crawler", "spider", "python", "curl", "wget"}
	for _, pattern := range botPatterns {
		if strings.Contains(strings.ToLower(userAgent), pattern) {
			return true, "Bot Traffic"
		}
	}
	return false, ""
}

func logTraffic(db *sqlx.DB, banService *services.DynamicBanService, notificationService *services.NotificationService, c *gin.Context, duration time.Duration) {
	// Detect attack
	var isAttack bool
	var attackType string

	// First check if downstream middleware (like OWASP) detected an attack
	if val, exists := c.Get("is_attack"); exists {
		isAttack = val.(bool)
	}
	if val, exists := c.Get("attack_type"); exists {
		attackType = val.(string)
	}

	// Fallback to internal detection if not already detected
	if !isAttack {
		isAttack, attackType = detectAttackType(c)
	}
	blocked := c.GetBool("blocked") || c.Writer.Status() == 403

	// If blocked due to attack, record violation
	if blocked && isAttack && banService != nil {
		realClientIP := GetRealClientIP(c)
		err := banService.RecordViolation(realClientIP, attackType)
		if err != nil {
			log.Printf("Failed to record violation for IP %s: %v", realClientIP, err)
		}
	}

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

		// Notify on Critical/High attacks
		notifyIfAttack(notificationService, c, isAttack, attackType, blockReason, db)
	}

	realClientIP := GetRealClientIP(c)
	countryCode := getCountryCode(realClientIP)

	// Get the HTTP Host header and try to find the matching vhost domain
	httpHost := c.Request.Host
	host := lookupVHostDomain(db, httpHost)

	_, err := db.Exec(query,
		time.Now(),
		realClientIP,
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

func notifyIfAttack(notificationService *services.NotificationService, c *gin.Context, isAttack bool, attackType, blockReason string, db *sqlx.DB) {
	if isAttack && notificationService != nil {
		// Get severity from context if available (set by OWASP middleware)
		severity := "medium"
		if val, exists := c.Get("attack_severity"); exists {
			severity = val.(string)
		}

		if severity == "critical" || severity == "high" {
			notificationService.Notify(services.NotificationEvent{
				Type:      "attack_detected",
				Title:     fmt.Sprintf("%s Detected", attackType),
				Message:   fmt.Sprintf("Blocked %s attack from IP %s on %s", attackType, GetRealClientIP(c), c.Request.Host),
				Severity:  severity,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"IP":     GetRealClientIP(c),
					"URL":    c.Request.URL.String(),
					"Host":   lookupVHostDomain(db, c.Request.Host),
					"Reason": blockReason,
				},
			})
		}
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
