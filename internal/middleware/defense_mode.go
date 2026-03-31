package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// DefenseModeMiddleware acts as the primary gatekeeper for the proxy.
// It sets the defense mode to context and handles offline immediately.
func DefenseModeMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		domain := c.Request.Host
		if colonIdx := strings.Index(domain, ":"); colonIdx != -1 {
			domain = domain[:colonIdx]
		}

		var defenseMode string
		err := db.Get(&defenseMode, "SELECT COALESCE(defense_mode, 'defense') FROM vhosts WHERE domain = $1", domain)
		if err != nil {
			defenseMode = "defense" // default if not found or DB error
		}

		c.Set("defense_mode", defenseMode)

		if defenseMode == "offline" {
			log.Printf("[DefenseMode] Domain %s is offline, returning 503", domain)
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusServiceUnavailable, getOfflinePageHTML(domain))
			c.Abort()
			return
		}

		c.Next()
	}
}

func getOfflinePageHTML(domain string) string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Website Offline - ` + domain + `</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; background: #f8fafc; color: #1e293b; display: flex; justify-content: center; align-items: center; min-height: 100vh; margin: 0; text-align: center; }
        .container { background: white; padding: 40px; border-radius: 20px; box-shadow: 0 10px 30px rgba(0,0,0,0.05); max-width: 500px; width: 90%; border-top: 5px solid #64748b; }
        h1 { color: #334155; margin-top: 0; font-size: 24px; }
        p { color: #64748b; font-size: 15px; line-height: 1.6; margin-bottom: 20px; }
        .icon { font-size: 64px; margin-bottom: 20px; }
        .footer { font-size: 12px; color: #94a3b8; margin-top: 30px; border-top: 1px solid #e2e8f0; padding-top: 15px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">⛔</div>
        <h1>Website Unavailable</h1>
        <p>The website <strong>` + domain + `</strong> is currently locked in <strong>Offline Mode</strong> by the Network Administrator.</p>
        <p>All traffic to this virtual host is temporarily suspended via the Web Application Firewall.</p>
        <div class="footer">Protected by DoCode WAF</div>
    </div>
</body>
</html>`
}
