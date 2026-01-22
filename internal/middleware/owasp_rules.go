package middleware

import (
	"log"
	"net/http"
	"regexp"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// OWASPRuleset contains compiled regex patterns for attack detection
type OWASPRuleset struct {
	SQLInjection     []*regexp.Regexp
	XSS              []*regexp.Regexp
	PathTraversal    []*regexp.Regexp
	CommandInjection []*regexp.Regexp
	LDAPInjection    []*regexp.Regexp
	XMLInjection     []*regexp.Regexp
	SSRF             []*regexp.Regexp
	Serialization    []*regexp.Regexp
	Log4Shell        []*regexp.Regexp
	ProtocolAttacks  []*regexp.Regexp
}

var (
	owaspRuleset     *OWASPRuleset
	owaspRulesetOnce sync.Once
)

// AttackResult contains detected attack information
type AttackResult struct {
	IsAttack    bool
	AttackType  string
	Severity    string // low, medium, high, critical
	Pattern     string
	Description string
}

// initOWASPRuleset initializes comprehensive OWASP ruleset
func initOWASPRuleset() {
	owaspRulesetOnce.Do(func() {
		owaspRuleset = &OWASPRuleset{}

		// SQL Injection Patterns (A03:2021 - Injection)
		sqlPatterns := []string{
			// Classic SQL injection
			`(?i)(\%27)|(')|(--)|(\%23)|(#)`,
			`(?i)((\%3D)|(=))[^\n]*((\%27)|(')|(--)|(\%3B)|(;))`,
			`(?i)\w*((\%27)|(\'))((\%6F)|o|(\%4F))((\%72)|r|(\%52))`,
			`(?i)union[\s\+]+select`,
			`(?i)union[\s\+]+all[\s\+]+select`,
			`(?i)select[\s\+]+[\w\*\(\)\,]+[\s\+]+from`,
			`(?i)insert[\s\+]+into[\s\+]+[\w]+`,
			`(?i)update[\s\+]+[\w]+[\s\+]+set`,
			`(?i)delete[\s\+]+from`,
			`(?i)drop[\s\+]+(table|database|index|view)`,
			`(?i)truncate[\s\+]+table`,
			`(?i)alter[\s\+]+table`,
			`(?i)create[\s\+]+(table|database|index|view)`,
			`(?i)exec[\s\+]+xp_`,
			`(?i)exec[\s\+]+sp_`,
			`(?i)xp_cmdshell`,
			`(?i)information_schema`,
			`(?i)sys\.tables`,
			`(?i)sys\.columns`,
			`(?i)@@version`,
			`(?i)@@servername`,
			`(?i)@@hostname`,
			`(?i)benchmark\s*\(`,
			`(?i)sleep\s*\(`,
			`(?i)waitfor[\s\+]+delay`,
			`(?i)having\s+[\d\w]+=[\d\w]+`,
			`(?i)group\s+by\s+[\d\w]+\s+having`,
			`(?i)order\s+by\s+\d+`,
			`(?i)load_file\s*\(`,
			`(?i)into\s+outfile`,
			`(?i)into\s+dumpfile`,
			`(?i)hex\s*\([\w]*\)`,
			`(?i)unhex\s*\(`,
			`(?i)char\s*\(\d+`,
			`(?i)concat\s*\(`,
			`(?i)concat_ws\s*\(`,
			`(?i)group_concat\s*\(`,
			`(?i)ascii\s*\(`,
			`(?i)ord\s*\(`,
			`(?i)substring\s*\(`,
			`(?i)mid\s*\(`,
			`(?i)extractvalue\s*\(`,
			`(?i)updatexml\s*\(`,
			// MySQL specific
			`(?i)procedure\s+analyse\s*\(`,
			// PostgreSQL specific
			`(?i)pg_sleep\s*\(`,
			`(?i)pg_user`,
			`(?i)current_database\s*\(`,
			// Oracle specific
			`(?i)utl_http`,
			`(?i)dbms_pipe`,
			// MSSQL specific
			`(?i)sp_executesql`,
			`(?i)openrowset`,
			`(?i)opendatasource`,
		}

		// XSS Patterns (A03:2021 - Injection)
		xssPatterns := []string{
			`(?i)<script[^>]*>`,
			`(?i)</script>`,
			`(?i)javascript\s*:`,
			`(?i)vbscript\s*:`,
			`(?i)on\w+\s*=`,
			`(?i)<img[^>]+onerror`,
			`(?i)<img[^>]+onload`,
			`(?i)<svg[^>]+onload`,
			`(?i)<body[^>]+onload`,
			`(?i)<iframe`,
			`(?i)<object`,
			`(?i)<embed`,
			`(?i)<form[^>]+action`,
			`(?i)<input[^>]+onfocus`,
			`(?i)<marquee`,
			`(?i)<bgsound`,
			`(?i)<link[^>]+href`,
			`(?i)<meta[^>]+http-equiv`,
			`(?i)<base[^>]+href`,
			`(?i)expression\s*\(`,
			`(?i)alert\s*\(`,
			`(?i)confirm\s*\(`,
			`(?i)prompt\s*\(`,
			`(?i)document\.cookie`,
			`(?i)document\.domain`,
			`(?i)document\.write`,
			`(?i)document\.location`,
			`(?i)window\.location`,
			`(?i)window\.open`,
			`(?i)\.innerHTML\s*=`,
			`(?i)\.outerHTML\s*=`,
			`(?i)eval\s*\(`,
			`(?i)Function\s*\(`,
			`(?i)setTimeout\s*\([^,]*['"]+`,
			`(?i)setInterval\s*\([^,]*['"]+`,
			`(?i)fromCharCode`,
			`(?i)String\.fromCodePoint`,
			`(?i)atob\s*\(`,
			`(?i)btoa\s*\(`,
			`(?i)data\s*:\s*text/html`,
			`(?i)data\s*:\s*image/svg`,
		}

		// Path Traversal Patterns (A01:2021 - Broken Access Control)
		pathTraversalPatterns := []string{
			`(?i)\.\.(/|\\)`,
			`(?i)\.\.%2f`,
			`(?i)\.\.%5c`,
			`(?i)%2e%2e/`,
			`(?i)%2e%2e\\`,
			`(?i)%252e%252e`,
			`(?i)/etc/passwd`,
			`(?i)/etc/shadow`,
			`(?i)/etc/hosts`,
			`(?i)/proc/self`,
			`(?i)/proc/version`,
			`(?i)c:\\windows\\system32`,
			`(?i)c:\\boot\.ini`,
			`(?i)c:\\inetpub`,
			`(?i)%00`,    // Null byte injection
			`(?i)%0d%0a`, // CRLF injection
		}

		// Command Injection Patterns (A03:2021 - Injection)
		cmdPatterns := []string{
			`(?i)[;&|]\s*(ls|dir|cat|type|more|less|head|tail|find)`,
			`(?i)[;&|]\s*(whoami|id|uname|hostname)`,
			`(?i)[;&|]\s*(nc|netcat|ncat|wget|curl|fetch)`,
			`(?i)[;&|]\s*(ping|traceroute|nslookup|dig)`,
			`(?i)[;&|]\s*(rm|del|erase|format)`,
			`(?i)[;&|]\s*(chmod|chown|chgrp)`,
			`(?i)[;&|]\s*(killall|pkill|kill)`,
			`(?i)\$\(.*\)`,
			"(?i)`[^`]*`",
			`(?i)\|\s*\w+`,
			`(?i)>\s*/dev/`,
			`(?i)>>\s*/`,
			`(?i)/bin/(ba)?sh`,
			`(?i)/bin/(c|tc|z|k)sh`,
			`(?i)cmd\.exe`,
			`(?i)powershell`,
			`(?i)python\s+-c`,
			`(?i)perl\s+-e`,
			`(?i)ruby\s+-e`,
			`(?i)php\s+-r`,
			`(?i)node\s+-e`,
		}

		// LDAP Injection Patterns (A03:2021 - Injection)
		ldapPatterns := []string{
			`(?i)\)\s*\(\|`,
			`(?i)\)\s*\(\&`,
			`(?i)\)\s*\(`,
			`(?i)\*\)\s*\(`,
			`(?i)\|\s*\(`,
			`(?i)&\s*\(`,
			`(?i)objectClass\s*=`,
			`(?i)objectCategory\s*=`,
			`(?i)userPassword\s*=`,
			`(?i)unicodePwd\s*=`,
		}

		// XML/XXE Injection Patterns (A03:2021 - Injection)
		xmlPatterns := []string{
			`(?i)<!ENTITY`,
			`(?i)<!DOCTYPE[^>]*\[`,
			`(?i)SYSTEM\s+["']`,
			`(?i)PUBLIC\s+["']`,
			`(?i)file://`,
			`(?i)php://`,
			`(?i)expect://`,
			`(?i)data://`,
			`(?i)gopher://`,
			`(?i)jar:file://`,
			`(?i)<\?xml[^>]*encoding`,
		}

		// SSRF Patterns (A10:2021 - Server-Side Request Forgery)
		ssrfPatterns := []string{
			`(?i)http://localhost`,
			`(?i)http://127\.0\.0\.1`,
			`(?i)http://\[::1\]`,
			`(?i)http://0\.0\.0\.0`,
			`(?i)http://169\.254\.169\.254`,                // AWS metadata
			`(?i)http://metadata\.google`,                  // GCP metadata
			`(?i)http://100\.100\.100\.200`,                // Alibaba metadata
			`(?i)http://192\.168\.\d+\.\d+`,                // Private network
			`(?i)http://10\.\d+\.\d+\.\d+`,                 // Private network
			`(?i)http://172\.(1[6-9]|2\d|3[01])\.\d+\.\d+`, // Private network
			`(?i)file://`,
			`(?i)dict://`,
			`(?i)gopher://`,
			`(?i)ftp://localhost`,
			`(?i)ldap://localhost`,
		}

		// Serialization Attack Patterns (A08:2021 - Software and Data Integrity Failures)
		serializationPatterns := []string{
			`(?i)O:\d+:"`,               // PHP object serialization
			`(?i)a:\d+:{`,               // PHP array serialization
			`(?i)s:\d+:"`,               // PHP string serialization
			`(?i)rO0ABX`,                // Java serialization (Base64)
			`(?i)aced0005`,              // Java serialization (Hex)
			`(?i){"@type"`,              // Fastjson
			`(?i){"_class"`,             // Spring Data
			`(?i)com\.sun\.org\.apache`, // Java gadget chains
			`(?i)java\.lang\.Runtime`,
			`(?i)java\.lang\.ProcessBuilder`,
			`(?i)org\.apache\.xalan`,
			`(?i)org\.springframework\.beans`,
			`(?i)__proto__`, // JavaScript prototype pollution
			`(?i)constructor\s*\[`,
		}

		// Log4Shell/JNDI Patterns (Critical CVE-2021-44228)
		log4shellPatterns := []string{
			`(?i)\$\{jndi:`,
			`(?i)\$\{lower:`,
			`(?i)\$\{upper:`,
			`(?i)\$\{env:`,
			`(?i)\$\{sys:`,
			`(?i)\$\{java:`,
			`(?i)\$\{base64:`,
			`(?i)\$\{date:`,
			`(?i)\$\{\$\{`,
			`(?i)%24%7bjndi`,
			`(?i)%24%7Bjndi`,
			`(?i)\$%7Bjndi`,
			`(?i)\$%7bjndi`,
			`(?i)ldap://`,
			`(?i)ldaps://`,
			`(?i)rmi://`,
			`(?i)dns://`,
		}

		// Protocol Attack Patterns
		protocolPatterns := []string{
			`(?i)%0d%0a`,           // CRLF injection
			`(?i)\r\n`,             // CRLF
			`(?i)Host:\s*[\w\.-]+`, // HTTP header injection
			`(?i)Content-Length:\s*\d+`,
			`(?i)Transfer-Encoding:\s*chunked`,
			`(?i)X-Forwarded-For:\s*\d`,
			`(?i)X-Original-URL:`,
			`(?i)X-Rewrite-URL:`,
		}

		// Compile all patterns
		owaspRuleset.SQLInjection = compilePatterns(sqlPatterns)
		owaspRuleset.XSS = compilePatterns(xssPatterns)
		owaspRuleset.PathTraversal = compilePatterns(pathTraversalPatterns)
		owaspRuleset.CommandInjection = compilePatterns(cmdPatterns)
		owaspRuleset.LDAPInjection = compilePatterns(ldapPatterns)
		owaspRuleset.XMLInjection = compilePatterns(xmlPatterns)
		owaspRuleset.SSRF = compilePatterns(ssrfPatterns)
		owaspRuleset.Serialization = compilePatterns(serializationPatterns)
		owaspRuleset.Log4Shell = compilePatterns(log4shellPatterns)
		owaspRuleset.ProtocolAttacks = compilePatterns(protocolPatterns)

		log.Printf("[OWASP] Ruleset initialized: %d SQL, %d XSS, %d PathTraversal, %d CmdInjection, %d LDAP, %d XML, %d SSRF, %d Serialization, %d Log4Shell, %d Protocol patterns",
			len(owaspRuleset.SQLInjection),
			len(owaspRuleset.XSS),
			len(owaspRuleset.PathTraversal),
			len(owaspRuleset.CommandInjection),
			len(owaspRuleset.LDAPInjection),
			len(owaspRuleset.XMLInjection),
			len(owaspRuleset.SSRF),
			len(owaspRuleset.Serialization),
			len(owaspRuleset.Log4Shell),
			len(owaspRuleset.ProtocolAttacks),
		)
	})
}

func compilePatterns(patterns []string) []*regexp.Regexp {
	compiled := make([]*regexp.Regexp, 0, len(patterns))
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			log.Printf("[OWASP] Failed to compile pattern: %s - %v", pattern, err)
			continue
		}
		compiled = append(compiled, re)
	}
	return compiled
}

// DetectOWASPAttack scans request for OWASP attack patterns
func DetectOWASPAttack(url, body, userAgent string, headers map[string]string) *AttackResult {
	initOWASPRuleset()

	// Combine all input for scanning
	input := strings.ToLower(url + " " + body + " " + userAgent)
	for k, v := range headers {
		input += " " + k + ": " + v
	}

	// Check Log4Shell first (Critical)
	for _, re := range owaspRuleset.Log4Shell {
		if re.MatchString(input) {
			return &AttackResult{
				IsAttack:    true,
				AttackType:  "Log4Shell/JNDI Injection",
				Severity:    "critical",
				Pattern:     re.String(),
				Description: "Potential Log4j remote code execution attempt (CVE-2021-44228)",
			}
		}
	}

	// Check Serialization Attacks (Critical)
	for _, re := range owaspRuleset.Serialization {
		if re.MatchString(input) {
			return &AttackResult{
				IsAttack:    true,
				AttackType:  "Serialization Attack",
				Severity:    "critical",
				Pattern:     re.String(),
				Description: "Potential deserialization attack detected",
			}
		}
	}

	// Check Command Injection (High)
	for _, re := range owaspRuleset.CommandInjection {
		if re.MatchString(input) {
			return &AttackResult{
				IsAttack:    true,
				AttackType:  "Command Injection",
				Severity:    "high",
				Pattern:     re.String(),
				Description: "OS command injection attempt detected",
			}
		}
	}

	// Check SQL Injection (High)
	for _, re := range owaspRuleset.SQLInjection {
		if re.MatchString(input) {
			return &AttackResult{
				IsAttack:    true,
				AttackType:  "SQL Injection",
				Severity:    "high",
				Pattern:     re.String(),
				Description: "SQL injection attempt detected",
			}
		}
	}

	// Check SSRF (High)
	for _, re := range owaspRuleset.SSRF {
		if re.MatchString(input) {
			return &AttackResult{
				IsAttack:    true,
				AttackType:  "SSRF",
				Severity:    "high",
				Pattern:     re.String(),
				Description: "Server-Side Request Forgery attempt detected",
			}
		}
	}

	// Check XSS (Medium)
	for _, re := range owaspRuleset.XSS {
		if re.MatchString(input) {
			return &AttackResult{
				IsAttack:    true,
				AttackType:  "XSS",
				Severity:    "medium",
				Pattern:     re.String(),
				Description: "Cross-Site Scripting attempt detected",
			}
		}
	}

	// Check Path Traversal (Medium)
	for _, re := range owaspRuleset.PathTraversal {
		if re.MatchString(input) {
			return &AttackResult{
				IsAttack:    true,
				AttackType:  "Path Traversal",
				Severity:    "medium",
				Pattern:     re.String(),
				Description: "Directory traversal attempt detected",
			}
		}
	}

	// Check LDAP Injection (Medium)
	for _, re := range owaspRuleset.LDAPInjection {
		if re.MatchString(input) {
			return &AttackResult{
				IsAttack:    true,
				AttackType:  "LDAP Injection",
				Severity:    "medium",
				Pattern:     re.String(),
				Description: "LDAP injection attempt detected",
			}
		}
	}

	// Check XML/XXE Injection (Medium)
	for _, re := range owaspRuleset.XMLInjection {
		if re.MatchString(input) {
			return &AttackResult{
				IsAttack:    true,
				AttackType:  "XML/XXE Injection",
				Severity:    "medium",
				Pattern:     re.String(),
				Description: "XML/XXE injection attempt detected",
			}
		}
	}

	// Check Protocol Attacks (Low)
	for _, re := range owaspRuleset.ProtocolAttacks {
		if re.MatchString(input) {
			return &AttackResult{
				IsAttack:    true,
				AttackType:  "Protocol Attack",
				Severity:    "low",
				Pattern:     re.String(),
				Description: "HTTP protocol manipulation attempt detected",
			}
		}
	}

	return &AttackResult{IsAttack: false}
}

// OWASPProtectionMiddleware provides comprehensive attack detection and blocking
func OWASPProtectionMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get current vhost domain
		domain := c.Request.Host
		if colonIdx := strings.Index(domain, ":"); colonIdx != -1 {
			domain = domain[:colonIdx]
		}

		// Get real client IP (from proxy headers)
		clientIP := GetRealClientIP(c)

		// Check if IP is whitelisted - skip OWASP protection for whitelisted IPs
		whitelisted, _ := isIPInGroup(db, clientIP, domain, "whitelist")
		if whitelisted {
			log.Printf("[OWASP] IP %s is whitelisted for domain %s - skipping OWASP protection", clientIP, domain)
			c.Next()
			return
		}

		// Check if OWASP protection is enabled for this vhost
		var vhostSettings struct {
			OWASPProtectionEnabled bool   `db:"owasp_protection_enabled"`
			OWASPProtectionLevel   string `db:"owasp_protection_level"`
		}

		err := db.Get(&vhostSettings, "SELECT COALESCE(owasp_protection_enabled, true) as owasp_protection_enabled, COALESCE(owasp_protection_level, 'medium') as owasp_protection_level FROM vhosts WHERE domain = $1", domain)
		if err != nil {
			// If vhost not found, use defaults
			vhostSettings.OWASPProtectionEnabled = true
			vhostSettings.OWASPProtectionLevel = "medium"
		}

		// If OWASP protection is disabled, skip
		if !vhostSettings.OWASPProtectionEnabled {
			c.Next()
			return
		}

		// Collect request data for scanning
		url := c.Request.URL.String()
		userAgent := c.GetHeader("User-Agent")

		// Get headers
		headers := make(map[string]string)
		for k, v := range c.Request.Header {
			if len(v) > 0 {
				headers[k] = v[0]
			}
		}

		// Read body if present (for POST/PUT)
		var body string
		if c.Request.Body != nil && (c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "PATCH") {
			// Note: Body reading should be done carefully to not consume the body
			// In production, consider using c.Request.GetBody() or buffering
			// For now, we scan URL and headers only to avoid body consumption issues
		}

		// Detect attack
		result := DetectOWASPAttack(url, body, userAgent, headers)

		if result.IsAttack {
			// Determine if we should block based on protection level
			shouldBlock := false
			switch vhostSettings.OWASPProtectionLevel {
			case "paranoid":
				// Block all attacks
				shouldBlock = true
			case "high":
				// Block critical, high, and medium severity
				shouldBlock = result.Severity != "low"
			case "medium":
				// Block critical and high severity
				shouldBlock = result.Severity == "critical" || result.Severity == "high"
			case "low":
				// Block only critical severity
				shouldBlock = result.Severity == "critical"
			}

			// Log the attack
			log.Printf("[OWASP] Attack detected: type=%s, severity=%s, domain=%s, ip=%s, url=%s, blocked=%v",
				result.AttackType, result.Severity, domain, clientIP, url, shouldBlock)

			// Set attack info in context for logging middleware
			c.Set("is_attack", true)
			c.Set("attack_type", result.AttackType)
			c.Set("attack_severity", result.Severity)

			if shouldBlock {
				c.Set("blocked", true)
				c.Set("block_reason", "OWASP: "+result.AttackType)
				c.Header("Content-Type", "text/html; charset=utf-8")
				c.String(http.StatusForbidden, getOWASPBlockedPageHTML(clientIP, domain, result))
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

// getOWASPBlockedPageHTML returns HTML for OWASP blocked page
func getOWASPBlockedPageHTML(clientIP, domain string, result *AttackResult) string {
	severityColor := "#f56565"
	severityBg := "#fff5f5"
	severityBorder := "#feb2b2"

	switch result.Severity {
	case "critical":
		severityColor = "#742a2a"
		severityBg = "#fed7d7"
		severityBorder = "#fc8181"
	case "high":
		severityColor = "#c53030"
		severityBg = "#fff5f5"
		severityBorder = "#feb2b2"
	case "medium":
		severityColor = "#c05621"
		severityBg = "#fffaf0"
		severityBorder = "#fbd38d"
	case "low":
		severityColor = "#2c7a7b"
		severityBg = "#e6fffa"
		severityBorder = "#81e6d9"
	}

	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Request Blocked - Security Alert</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, ` + severityColor + ` 0%, #2d3748 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.3);
            max-width: 550px;
            width: 100%;
            padding: 40px;
            text-align: center;
        }
        .icon {
            width: 100px; height: 100px;
            margin: 0 auto 20px;
            background: linear-gradient(135deg, ` + severityColor + ` 0%, #2d3748 100%);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .icon svg { width: 50px; height: 50px; fill: white; }
        h1 { color: #2d3748; font-size: 28px; margin-bottom: 10px; }
        .subtitle { color: #718096; margin-bottom: 25px; }
        .alert-box {
            background: ` + severityBg + `;
            border: 2px solid ` + severityBorder + `;
            border-radius: 15px;
            padding: 20px;
            margin: 20px 0;
            text-align: left;
        }
        .alert-title { color: ` + severityColor + `; font-weight: 700; font-size: 16px; margin-bottom: 10px; }
        .alert-item { color: #4a5568; font-size: 14px; padding: 8px 0; border-bottom: 1px solid ` + severityBorder + `; display: flex; justify-content: space-between; }
        .alert-item:last-child { border-bottom: none; }
        .alert-label { font-weight: 600; color: ` + severityColor + `; }
        .severity-badge {
            display: inline-block;
            padding: 4px 12px;
            border-radius: 9999px;
            font-size: 12px;
            font-weight: 700;
            text-transform: uppercase;
            background: ` + severityColor + `;
            color: white;
        }
        .footer { margin-top: 30px; color: #a0aec0; font-size: 14px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">
            <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
            </svg>
        </div>
        <h1>🛡️ Request Blocked</h1>
        <p class="subtitle">Your request has been blocked by the security system</p>
        
        <div class="alert-box">
            <div class="alert-title">⚠️ Security Alert Details</div>
            <div class="alert-item">
                <span class="alert-label">Attack Type</span>
                <span>` + result.AttackType + `</span>
            </div>
            <div class="alert-item">
                <span class="alert-label">Severity</span>
                <span class="severity-badge">` + result.Severity + `</span>
            </div>
            <div class="alert-item">
                <span class="alert-label">Your IP</span>
                <span>` + clientIP + `</span>
            </div>
            <div class="alert-item">
                <span class="alert-label">Domain</span>
                <span>` + domain + `</span>
            </div>
        </div>
        
        <p style="color: #718096; font-size: 14px; margin: 20px 0;">
            If you believe this is a mistake, please contact the site administrator.
        </p>
        
        <p class="footer">🛡️ Protected by DoCode WAF - OWASP Top 10 Protection</p>
    </div>
</body>
</html>`
}
