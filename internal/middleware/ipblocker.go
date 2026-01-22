package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// GetRealClientIP extracts the real client IP from proxy headers
// This is essential when running behind nginx or other reverse proxies
func GetRealClientIP(c *gin.Context) string {
	// Try to get real client IP from X-Forwarded-For or X-Real-IP headers
	clientIP := c.ClientIP()
	xForwardedFor := c.GetHeader("X-Forwarded-For")
	xRealIP := c.GetHeader("X-Real-IP")

	// Use X-Real-IP if available (set by nginx/proxy)
	if xRealIP != "" {
		clientIP = xRealIP
	} else if xForwardedFor != "" {
		// X-Forwarded-For can contain multiple IPs, take the first one
		ips := strings.Split(xForwardedFor, ",")
		if len(ips) > 0 {
			clientIP = strings.TrimSpace(ips[0])
		}
	}

	return clientIP
}

// IPBlockerMiddleware blocks requests from blacklisted IPs and allows only whitelisted IPs
func IPBlockerMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get real client IP using helper function
		clientIP := GetRealClientIP(c)
		xForwardedFor := c.GetHeader("X-Forwarded-For")
		xRealIP := c.GetHeader("X-Real-IP")

		// Get current vhost domain
		domain := c.Request.Host
		// Remove port if present
		if colonIdx := strings.Index(domain, ":"); colonIdx != -1 {
			domain = domain[:colonIdx]
		}

		// Debug logging
		log.Printf("[IP Blocker] Client IP: %s, Domain: %s, X-Forwarded-For: %s, X-Real-IP: %s, RemoteAddr: %s",
			clientIP, domain, xForwardedFor, xRealIP, c.Request.RemoteAddr)

		// Check if vhost has an active whitelist
		hasWhitelist, err := vhostHasActiveWhitelist(db, domain)
		if err != nil {
			log.Printf("[IP Blocker] Error checking whitelist for domain %s: %v", domain, err)
		}

		// Check whitelist first (both global and vhost-specific)
		whitelisted, err := isIPInGroup(db, clientIP, domain, "whitelist")
		if err == nil && whitelisted {
			log.Printf("[IP Blocker] IP %s is whitelisted for domain %s", clientIP, domain)
			c.Next()
			return
		}

		// If vhost has an active whitelist and IP is not in it, block the request
		if hasWhitelist {
			log.Printf("[IP Blocker] IP %s is NOT in whitelist for domain %s - blocking request (whitelist mode)", clientIP, domain)
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusForbidden, getWhitelistBlockedPageHTML(clientIP, domain))
			c.Abort()
			return
		}

		// Check blacklist (both global and vhost-specific)
		blacklisted, err := isIPInGroup(db, clientIP, domain, "blacklist")
		if err == nil && blacklisted {
			log.Printf("[IP Blocker] IP %s is blacklisted for domain %s - blocking request", clientIP, domain)
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusForbidden, getBlockedPageHTML(db, clientIP, c.Request.Host))
			c.Abort()
			return
		}

		// Check blocking rules
		blocked, reason := checkBlockingRules(db, c)
		if blocked {
			c.JSON(http.StatusForbidden, gin.H{
				"error": reason,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// vhostHasActiveWhitelist checks if a domain has any active whitelist groups
func vhostHasActiveWhitelist(db *sqlx.DB, domain string) (bool, error) {
	query := `
		SELECT COUNT(*) > 0
		FROM ip_groups ig
		JOIN ip_group_vhosts igv ON ig.id = igv.ip_group_id
		JOIN vhosts v ON igv.vhost_id = v.id
		WHERE ig.type = 'whitelist' AND v.domain = $1
	`
	var hasWhitelist bool
	err := db.Get(&hasWhitelist, query, domain)
	if err != nil {
		return false, err
	}

	// Also check global whitelists (no vhost associations)
	if !hasWhitelist {
		globalQuery := `
			SELECT COUNT(*) > 0
			FROM ip_groups ig
			WHERE ig.type = 'whitelist' 
			AND NOT EXISTS (SELECT 1 FROM ip_group_vhosts WHERE ip_group_id = ig.id)
		`
		err = db.Get(&hasWhitelist, globalQuery)
		if err != nil {
			return false, err
		}
	}

	return hasWhitelist, nil
}

// getWhitelistBlockedPageHTML returns a styled HTML page for whitelist-blocked users
func getWhitelistBlockedPageHTML(clientIP, domain string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Access Restricted - Not in Whitelist</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #4f46e5 0%%, #7c3aed 100%%);
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
            max-width: 600px;
            width: 100%%;
            padding: 40px;
            text-align: center;
        }
        .icon {
            width: 80px;
            height: 80px;
            margin: 0 auto 25px;
            background: linear-gradient(135deg, #6366f1 0%%, #8b5cf6 100%%);
            border-radius: 50%%;
            display: flex;
            align-items: center;
            justify-content: center;
        }
        .icon svg { width: 50px; height: 50px; fill: white; }
        h1 { color: #1f2937; font-size: 28px; margin-bottom: 15px; }
        .subtitle { color: #6b7280; font-size: 16px; margin-bottom: 25px; line-height: 1.6; }
        .info-box {
            background: #f3f4f6;
            border-radius: 12px;
            padding: 20px;
            margin: 20px 0;
            text-align: left;
        }
        .info-row { display: flex; justify-content: space-between; margin-bottom: 10px; }
        .info-label { color: #6b7280; font-size: 14px; }
        .info-value { color: #1f2937; font-weight: 600; font-family: monospace; }
        .message {
            background: #eff6ff;
            border-left: 4px solid #3b82f6;
            padding: 15px;
            border-radius: 8px;
            text-align: left;
            margin-top: 20px;
        }
        .message p { color: #1e40af; line-height: 1.6; font-size: 14px; }
        .footer { margin-top: 25px; color: #9ca3af; font-size: 12px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">
            <svg viewBox="0 0 24 24"><path d="M12 1L3 5v6c0 5.55 3.84 10.74 9 12 5.16-1.26 9-6.45 9-12V5l-9-4zm0 10.99h7c-.53 4.12-3.28 7.79-7 8.94V12H5V6.3l7-3.11v8.8z"/></svg>
        </div>
        <h1>Access Restricted</h1>
        <p class="subtitle">This resource requires IP whitelist authorization.<br>Your IP address is not in the allowed list.</p>
        <div class="info-box">
            <div class="info-row">
                <span class="info-label">Your IP Address</span>
                <span class="info-value">%s</span>
            </div>
            <div class="info-row">
                <span class="info-label">Requested Domain</span>
                <span class="info-value">%s</span>
            </div>
        </div>
        <div class="message">
            <p>This domain uses <strong>IP Whitelist</strong> access control. Only pre-approved IP addresses can access this resource. Please contact your System Administrator to request access.</p>
        </div>
        <div class="footer">Protected by DoCode WAF</div>
    </div>
</body>
</html>`, clientIP, domain)
}

func isIPInGroup(db *sqlx.DB, clientIP, domain, groupType string) (bool, error) {
	// Check both global rules (no vhost associations) and vhost-specific rules
	// Updated to use ip_group_vhosts junction table
	query := `
		SELECT DISTINCT ia.ip_address, ia.cidr_mask 
		FROM ip_addresses ia
		JOIN ip_groups ig ON ia.group_id = ig.id
		LEFT JOIN ip_group_vhosts igv ON ig.id = igv.ip_group_id
		LEFT JOIN vhosts v ON igv.vhost_id = v.id
		WHERE ig.type = $1 
		AND (
			-- Global rules: no vhost associations
			NOT EXISTS (SELECT 1 FROM ip_group_vhosts WHERE ip_group_id = ig.id)
			OR 
			-- Vhost-specific rules: matches current domain
			v.domain = $2
		)
	`

	var addresses []struct {
		IPAddress string `db:"ip_address"`
		CIDRMask  *int   `db:"cidr_mask"`
	}

	err := db.Select(&addresses, query, groupType, domain)
	if err != nil {
		log.Printf("[IP Blocker] Error querying %s for domain %s: %v", groupType, domain, err)
		return false, err
	}

	log.Printf("[IP Blocker] Checking client IP %s against %d %s entries for domain %s", clientIP, len(addresses), groupType, domain)

	clientIPParsed := net.ParseIP(clientIP)
	if clientIPParsed == nil {
		log.Printf("[IP Blocker] Failed to parse client IP: %s", clientIP)
		return false, nil
	}

	for _, addr := range addresses {
		// Check if CIDR is specified
		if addr.CIDRMask != nil && *addr.CIDRMask > 0 {
			// Build CIDR notation
			cidr := fmt.Sprintf("%s/%d", addr.IPAddress, *addr.CIDRMask)
			log.Printf("[IP Blocker] Checking CIDR: %s", cidr)

			_, ipNet, err := net.ParseCIDR(cidr)
			if err == nil && ipNet.Contains(clientIPParsed) {
				log.Printf("[IP Blocker] MATCH! IP %s is in CIDR %s", clientIP, cidr)
				return true, nil
			}
		} else {
			// Exact IP match
			log.Printf("[IP Blocker] Checking exact IP: %s", addr.IPAddress)
			if addr.IPAddress == clientIP {
				log.Printf("[IP Blocker] MATCH! IP %s matches exactly", clientIP)
				return true, nil
			}
		}
	}

	log.Printf("[IP Blocker] No match found for IP %s in %s", clientIP, groupType)
	return false, nil
}

func checkBlockingRules(db *sqlx.DB, c *gin.Context) (bool, string) {
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	url := c.Request.URL.Path

	var rules []struct {
		Type    string
		Pattern string
		Action  string
	}

	query := `
		SELECT type, pattern, action FROM blocking_rules 
		WHERE enabled = true 
		ORDER BY priority DESC
	`
	err := db.Select(&rules, query)
	if err != nil {
		return false, ""
	}

	for _, rule := range rules {
		matched := false

		switch rule.Type {
		case "ip":
			if matchIP(clientIP, rule.Pattern) {
				matched = true
			}
		case "url":
			if strings.Contains(url, rule.Pattern) {
				matched = true
			}
		case "user_agent":
			if strings.Contains(strings.ToLower(userAgent), strings.ToLower(rule.Pattern)) {
				matched = true
			}
		}

		if matched && rule.Action == "block" {
			return true, "Blocked by security rule"
		}
	}

	return false, ""
}

func matchIP(ip, pattern string) bool {
	// Check if pattern is CIDR
	if strings.Contains(pattern, "/") {
		_, ipNet, err := net.ParseCIDR(pattern)
		if err != nil {
			return false
		}
		testIP := net.ParseIP(ip)
		return ipNet.Contains(testIP)
	}
	// Exact match
	return ip == pattern
}

// getBlockedPageHTML returns a styled HTML page for blocked users
func getBlockedPageHTML(db *sqlx.DB, clientIP, host string) string {
	// Get application name from vhost
	// var appName string
	// query := "SELECT name FROM vhosts WHERE domain = $1 LIMIT 1"
	// err := db.Get(&appName, query, host)
	// if err != nil || appName == "" {
	// 	appName = "Web Application Firewall"
	// }
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Access Denied - IP Blocked</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #bf6715ff 0%, #991b1b 100%);
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
            max-width: 1200px;
            width: 100%;
            padding: 40px;
            text-align: center;
            animation: slideIn 0.5s ease-out;
        }
        
        @keyframes slideIn {
            from {
                opacity: 0;
                transform: translateY(-30px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
        
        .icon {
            width: 100px;
            height: 100px;
            margin: 0 auto 30px;
            background: linear-gradient(135deg, #ef4444 0%, #dc2626 100%);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            box-shadow: 0 10px 30px rgba(220, 38, 38, 0.4);
        }
        
        .icon svg {
            width: 60px;
            height: 60px;
            fill: white;
        }
        
        h1 {
            color: #2d3748;
            font-size: 32px;
            margin-bottom: 15px;
            font-weight: 700;
        }
        
        .subtitle {
            color: #718096;
            font-size: 18px;
            margin-bottom: 30px;
        }
        
        .ip-box {
            background: #f7fafc;
            border: 2px solid #e2e8f0;
            border-radius: 10px;
            padding: 15px;
            margin: 25px auto;
            display: inline-block;
            max-width: 300px;
        }
        
        .ip-label {
            color: #718096;
            font-size: 14px;
            margin-bottom: 5px;
        }
        
        .ip-address {
            color: #2d3748;
            font-size: 20px;
            font-weight: 600;
            font-family: 'Courier New', monospace;
        }

        .info-boxes {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 20px;
            margin: 25px 0;
            text-align: left;
        }
        
        .message-box {
            background: #fff5f5;
            border-left: 4px solid #f56565;
            padding: 20px;
            border-radius: 8px;
        }
        
        .message-box h3 {
            color: #c53030;
            font-size: 18px;
            margin-bottom: 10px;
            display: flex;
            align-items: center;
        }
        
        .message-box h3 svg {
            width: 20px;
            height: 20px;
            margin-right: 8px;
            fill: #c53030;
        }
        
        .message-box p {
            color: #742a2a;
            line-height: 1.6;
            margin-bottom: 8px;
        }
        
        .contact-box {
            background: #ebf8ff;
            border-left: 4px solid #4299e1;
            padding: 20px;
            border-radius: 8px;
        }
        
        .contact-box h3 {
            color: #2c5282;
            font-size: 18px;
            margin-bottom: 10px;
            display: flex;
            align-items: center;
        }
        
        .contact-box h3 svg {
            width: 20px;
            height: 20px;
            margin-right: 8px;
            fill: #2c5282;
        }
        
        .contact-box p {
            color: #2c5282;
            line-height: 1.6;
        }
        
        .contact-box ul {
            margin-top: 15px;
            padding-left: 20px;
        }
        
        .contact-box li {
            color: #2c5282;
            margin-bottom: 8px;
            line-height: 1.6;
        }
        
        .contact-box strong {
            color: #1a365d;
        }
        
        .footer {
            margin-top: 30px;
            padding-top: 20px;
            border-top: 1px solid #e2e8f0;
            color: #a0aec0;
            font-size: 14px;
        }
        
        @media (max-width: 600px) {
            .container {
                padding: 30px 20px;
            }
            
            h1 {
                font-size: 26px;
            }
            
            .subtitle {
                font-size: 16px;
            }

            .info-boxes {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">
            <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1 15h-2v-2h2v2zm0-4h-2V7h2v6z"/>
            </svg>
        </div>
        
        <h1>Access Denied</h1>
        <p class="subtitle">Your IP address has been blocked by our Web Application Firewall</p>
        
        <div class="ip-box">
            <div class="ip-label">Your IP Address</div>
            <div class="ip-address">` + clientIP + `</div>
        </div>
        
        <div class="info-boxes">
            <div class="message-box">
                <h3>
                    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                        <path d="M1 21h22L12 2 1 21zm12-3h-2v-2h2v2zm0-4h-2v-4h2v4z"/>
                    </svg>
                    Why am I seeing this?
                </h3>
                <p>Your IP address is currently on our <strong>blacklist</strong> due to security policies or suspicious activity detected from your network.</p>
                <p>Access to this resource has been restricted to protect our systems and other users.</p>
            </div>
            
            <div class="contact-box">
                <h3>
                    <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                        <path d="M20 4H4c-1.1 0-1.99.9-1.99 2L2 18c0 1.1.9 2 2 2h16c1.1 0 2-.9 2-2V6c0-1.1-.9-2-2-2zm0 4l-8 5-8-5V6l8 5 8-5v2z"/>
                    </svg>
                    Need Access?
                </h3>
                <p>If you believe this is a mistake or you need access to this resource, please contact your <strong>System Administrator</strong> with the following information:</p>
                <ul>
                    <li>Your IP address: <strong>` + clientIP + `</strong></li>
                    <li>Date and time of access attempt</li>
                    <li>Reason for access request</li>
                </ul>
                <p style="margin-top: 15px;">The System Administrator can review your request and move your IP from the blacklist to the whitelist if approved.</p>
            </div>
        </div>
        
        <div class="footer">
            <p>Protected by DoCode WAF (Web Application Firewall)</p>
        </div>
    </div>
</body>
</html>`
}
