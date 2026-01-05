package middleware

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/aleh/docode-waf/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// RegionFilter middleware checks if request is from allowed/blocked region
func RegionFilter(db *sqlx.DB, geoIPService *services.GeoIPService) gin.HandlerFunc {
	return func(c *gin.Context) {
		domain := c.Request.Host

		// Get vhost settings
		var vhostSettings struct {
			RegionWhitelist        []string `db:"region_whitelist"`
			RegionBlacklist        []string `db:"region_blacklist"`
			RegionFilteringEnabled bool     `db:"region_filtering_enabled"`
		}

		query := `
			SELECT 
				COALESCE(region_whitelist, '{}') as region_whitelist,
				COALESCE(region_blacklist, '{}') as region_blacklist,
				COALESCE(region_filtering_enabled, false) as region_filtering_enabled
			FROM vhosts 
			WHERE domain = $1 AND enabled = true
		`

		err := db.Get(&vhostSettings, query, domain)
		if err != nil {
			log.Printf("[Region Filter] Error getting vhost settings for domain %s: %v", domain, err)
			c.Next()
			return
		}

		// Skip if region filtering is disabled
		if !vhostSettings.RegionFilteringEnabled {
			c.Next()
			return
		}

		// Get client IP
		clientIP := getClientIP(c.Request)
		log.Printf("[Region Filter] Checking IP %s for domain %s", clientIP, domain)

		// Lookup country code
		countryCode, err := geoIPService.GetCountryCode(clientIP)
		if err != nil {
			log.Printf("[Region Filter] Failed to lookup IP %s: %v", clientIP, err)
			// On error, allow the request (fail open)
			c.Next()
			return
		}

		log.Printf("[Region Filter] IP %s resolved to country: %s", clientIP, countryCode)

		// Check whitelist first (if not empty, only whitelist countries are allowed)
		if len(vhostSettings.RegionWhitelist) > 0 {
			allowed := false
			for _, allowedCountry := range vhostSettings.RegionWhitelist {
				if allowedCountry == countryCode {
					allowed = true
					break
				}
			}
			if !allowed {
				log.Printf("[Region Filter] Blocked IP %s from country %s (not in whitelist)", clientIP, countryCode)
				c.HTML(http.StatusForbidden, "", getRegionBlockedPageHTML(domain, countryCode, "whitelist"))
				c.Abort()
				return
			}
		}

		// Check blacklist (if whitelist is empty or passed)
		if len(vhostSettings.RegionBlacklist) > 0 {
			for _, blockedCountry := range vhostSettings.RegionBlacklist {
				if blockedCountry == countryCode {
					log.Printf("[Region Filter] Blocked IP %s from blacklisted country %s", clientIP, countryCode)
					c.HTML(http.StatusForbidden, "", getRegionBlockedPageHTML(domain, countryCode, "blacklist"))
					c.Abort()
					return
				}
			}
		}

		log.Printf("[Region Filter] Allowed IP %s from country %s", clientIP, countryCode)
		c.Next()
	}
}

// getRegionBlockedPageHTML returns HTML for region-blocked page
func getRegionBlockedPageHTML(domain, countryCode, listType string) string {
	reason := "not in the allowed regions"
	if listType == "blacklist" {
		reason = "from a blocked region"
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Access Denied - Region Blocked</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
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
            max-width: 500px;
            width: 100%%;
            padding: 40px;
            text-align: center;
        }
        .icon {
            width: 80px;
            height: 80px;
            background: linear-gradient(135deg, #f093fb 0%%, #f5576c 100%%);
            border-radius: 50%%;
            display: flex;
            align-items: center;
            justify-content: center;
            margin: 0 auto 30px;
            font-size: 40px;
        }
        h1 {
            color: #2d3748;
            font-size: 28px;
            margin-bottom: 15px;
            font-weight: 700;
        }
        .message {
            color: #4a5568;
            font-size: 16px;
            line-height: 1.6;
            margin-bottom: 25px;
        }
        .info {
            background: #f7fafc;
            border-radius: 10px;
            padding: 20px;
            margin-bottom: 25px;
        }
        .info-item {
            display: flex;
            justify-content: space-between;
            padding: 8px 0;
            border-bottom: 1px solid #e2e8f0;
        }
        .info-item:last-child {
            border-bottom: none;
        }
        .info-label {
            color: #718096;
            font-weight: 500;
        }
        .info-value {
            color: #2d3748;
            font-weight: 600;
        }
        .footer {
            color: #a0aec0;
            font-size: 14px;
            margin-top: 20px;
        }
        @media (max-width: 600px) {
            .container {
                padding: 30px 20px;
            }
            h1 {
                font-size: 24px;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">🌍</div>
        <h1>Access Denied</h1>
        <p class="message">
            Sorry, access to <strong>%s</strong> is not available from your region.
        </p>
        <div class="info">
            <div class="info-item">
                <span class="info-label">Your Country:</span>
                <span class="info-value">%s</span>
            </div>
            <div class="info-item">
                <span class="info-label">Reason:</span>
                <span class="info-value">%s</span>
            </div>
        </div>
        <p class="footer">
            If you believe this is an error, please contact the website administrator.
        </p>
    </div>
</body>
</html>`, domain, countryCode, reason)
}

// getClientIP extracts the real client IP from request
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		return strings.TrimSpace(ips[0])
	}

	// Check X-Real-IP header
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return xri
	}

	// Fall back to RemoteAddr
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
