package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/aleh/docode-waf/internal/services"
	"github.com/gin-gonic/gin"
)

// DynamicBanMiddleware checks if the client IP is dynamically banned
func DynamicBanMiddleware(banService *services.DynamicBanService) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := GetRealClientIP(c)

		// Check if banned
		isBanned, reason := banService.IsBanned(clientIP)
		if isBanned {
			log.Printf("[DynamicBan] Blocked request from banned IP %s. Reason: %s", clientIP, reason)

			// Get domain for display in blocked page
			domain := c.Request.Host
			if idx := strings.Index(domain, ":"); idx != -1 {
				domain = domain[:idx]
			}

			// Return blocked page
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusForbidden, getDynamicBanPageHTML(clientIP, domain, reason))
			c.Abort()
			return
		}

		c.Next()
	}
}

func getDynamicBanPageHTML(clientIP, domain, reason string) string {
	// Re-using the style from OWASP page for consistency, but simplified
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Access Temporarily Suspended</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #1a202c 0%, #2d3748 100%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            padding: 20px;
            color: #e2e8f0;
        }
        .container {
            background: #2d3748;
            border-radius: 20px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.5);
            max-width: 500px;
            width: 100%;
            padding: 40px;
            text-align: center;
            border: 1px solid #4a5568;
        }
        .icon {
            font-size: 64px;
            margin-bottom: 20px;
        }
        h1 { font-size: 24px; margin-bottom: 10px; color: #fc8181; }
        p { margin-bottom: 20px; color: #cbd5e0; line-height: 1.6; }
        .reason-box {
            background: #742a2a;
            color: #fed7d7;
            padding: 15px;
            border-radius: 8px;
            font-size: 14px;
            margin-bottom: 20px;
            border: 1px solid #9b2c2c;
        }
        .ip-info { font-size: 12px; color: #718096; margin-top: 20px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">⛔</div>
        <h1>Temporarily Banned</h1>
        <p>Your access has been temporarily suspended due to suspicious activity detected from your IP address.</p>
        
        <div class="reason-box">
            <strong>Reason:</strong><br>
            ` + reason + `
        </div>
        
        <p style="font-size: 13px; color: #a0aec0;">
            The ban will expire automatically. Please try again later.
        </p>
        
        <div class="ip-info">
            IP: ` + clientIP + ` • Domain: ` + domain + `
        </div>
    </div>
</body>
</html>`
}
