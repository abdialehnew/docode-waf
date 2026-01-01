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

// IPBlockerMiddleware blocks requests from blacklisted IPs and allows only whitelisted IPs
func IPBlockerMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
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

		// Debug logging
		log.Printf("[IP Blocker] Client IP: %s, X-Forwarded-For: %s, X-Real-IP: %s, RemoteAddr: %s",
			clientIP, xForwardedFor, xRealIP, c.Request.RemoteAddr)

		// Check whitelist first
		whitelisted, err := isIPInGroup(db, clientIP, "whitelist")
		if err == nil && whitelisted {
			log.Printf("[IP Blocker] IP %s is whitelisted", clientIP)
			c.Next()
			return
		}

		// Check blacklist
		blacklisted, err := isIPInGroup(db, clientIP, "blacklist")
		if err == nil && blacklisted {
			log.Printf("[IP Blocker] IP %s is blacklisted - blocking request", clientIP)
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

func isIPInGroup(db *sqlx.DB, clientIP, groupType string) (bool, error) {
	query := `
		SELECT ia.ip_address, ia.cidr_mask 
		FROM ip_addresses ia
		JOIN ip_groups ig ON ia.group_id = ig.id
		WHERE ig.type = $1
	`

	var addresses []struct {
		IPAddress string `db:"ip_address"`
		CIDRMask  *int   `db:"cidr_mask"`
	}

	err := db.Select(&addresses, query, groupType)
	if err != nil {
		log.Printf("[IP Blocker] Error querying %s: %v", groupType, err)
		return false, err
	}

	log.Printf("[IP Blocker] Checking client IP %s against %d %s entries", clientIP, len(addresses), groupType)

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
            max-width: 600px;
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
            margin: 25px 0;
            display: inline-block;
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
        
        .message-box {
            background: #fff5f5;
            border-left: 4px solid #f56565;
            padding: 20px;
            margin: 25px 0;
            text-align: left;
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
            margin: 25px 0;
            text-align: left;
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
        
        <div class="footer">
            <p>Protected by DoCode WAF (Web Application Firewall)</p>
        </div>
    </div>
</body>
</html>`
}
