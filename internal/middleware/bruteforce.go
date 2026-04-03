package middleware

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

// BruteForceMiddleware implements brute force protection for login endpoints
// Tracks failed login attempts and progressively delays/blocks attackers
func BruteForceMiddleware(redisClient *redis.Client, db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip non-POST requests (logins are typically POST)
		if c.Request.Method != "POST" {
			c.Next()
			return
		}

		// Get current vhost domain
		domain := c.Request.Host
		if colonIdx := strings.Index(domain, ":"); colonIdx != -1 {
			domain = domain[:colonIdx]
		}

		// Check if brute force protection is enabled for this vhost
		var vhostSettings struct {
			BruteForceEnabled   bool   `db:"brute_force_enabled"`
			BruteForceThreshold int    `db:"brute_force_threshold"`
			BruteForceWindow    int    `db:"brute_force_window"`
			BruteForceLockout   int    `db:"brute_force_lockout"`
			LoginPaths          string `db:"login_paths"`
		}

		err := db.Get(&vhostSettings, `
			SELECT 
				COALESCE(brute_force_enabled, false) as brute_force_enabled,
				COALESCE(brute_force_threshold, 5) as brute_force_threshold,
				COALESCE(brute_force_window, 300) as brute_force_window,
				COALESCE(brute_force_lockout, 900) as brute_force_lockout,
				COALESCE(login_paths, '/login,/auth,/signin,/api/login,/api/auth') as login_paths
			FROM vhosts WHERE domain = $1`, domain)

		if err != nil {
			// If vhost not found, skip brute force protection
			c.Next()
			return
		}

		// If brute force protection is disabled, skip
		if !vhostSettings.BruteForceEnabled {
			c.Next()
			return
		}

		// Check if current path is a login path
		currentPath := strings.ToLower(c.Request.URL.Path)
		loginPaths := strings.Split(vhostSettings.LoginPaths, ",")
		isLoginPath := false
		for _, path := range loginPaths {
			path = strings.TrimSpace(strings.ToLower(path))
			if path != "" && strings.HasPrefix(currentPath, path) {
				isLoginPath = true
				break
			}
		}

		if !isLoginPath {
			c.Next()
			return
		}

		clientIP := GetRealClientIP(c)
		ctx := context.Background()

		// Keys for tracking
		attemptKey := fmt.Sprintf("bruteforce:attempts:%s:%s", domain, clientIP)
		lockoutKey := fmt.Sprintf("bruteforce:lockout:%s:%s", domain, clientIP)

		// Check if IP is locked out
		locked, err := redisClient.Exists(ctx, lockoutKey).Result()
		if err == nil && locked > 0 {
			ttl, _ := redisClient.TTL(ctx, lockoutKey).Result()
			log.Printf("[BruteForce] IP %s is locked out for domain %s, remaining: %v", clientIP, domain, ttl)
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Header("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))
			c.String(http.StatusTooManyRequests, getBruteForceBlockedHTML(clientIP, domain, int(ttl.Seconds())))
			c.Abort()
			return
		}

		// Get current attempt count
		attempts, err := redisClient.Get(ctx, attemptKey).Int()
		if err != nil && err != redis.Nil {
			// Redis error, continue without protection
			c.Next()
			return
		}

		// If threshold exceeded, lock out the IP
		if attempts >= vhostSettings.BruteForceThreshold {
			// Set lockout
			redisClient.Set(ctx, lockoutKey, "1", time.Duration(vhostSettings.BruteForceLockout)*time.Second)
			// Delete attempt counter
			redisClient.Del(ctx, attemptKey)

			log.Printf("[BruteForce] IP %s locked out for domain %s (exceeded %d attempts)",
				clientIP, domain, vhostSettings.BruteForceThreshold)

			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Header("Retry-After", fmt.Sprintf("%d", vhostSettings.BruteForceLockout))
			c.String(http.StatusTooManyRequests, getBruteForceBlockedHTML(clientIP, domain, vhostSettings.BruteForceLockout))
			c.Abort()
			return
		}

		// Process the request
		c.Next()

		// After request, check if login failed (4xx status code)
		statusCode := c.Writer.Status()
		if statusCode >= 400 && statusCode < 500 {
			// Increment failed attempt counter
			pipe := redisClient.Pipeline()
			pipe.Incr(ctx, attemptKey)
			pipe.Expire(ctx, attemptKey, time.Duration(vhostSettings.BruteForceWindow)*time.Second)
			pipe.Exec(ctx)

			newAttempts := attempts + 1
			remaining := vhostSettings.BruteForceThreshold - newAttempts
			log.Printf("[BruteForce] Failed login attempt from %s for domain %s (%d/%d, %d remaining)",
				clientIP, domain, newAttempts, vhostSettings.BruteForceThreshold, remaining)
		}
	}
}

// getBruteForceBlockedHTML returns HTML for brute force lockout page
func getBruteForceBlockedHTML(clientIP, domain string, lockoutSeconds int) string {
	minutes := lockoutSeconds / 60
	seconds := lockoutSeconds % 60
	timeDisplay := ""
	if minutes > 0 {
		timeDisplay = fmt.Sprintf("%d min %d sec", minutes, seconds)
	} else {
		timeDisplay = fmt.Sprintf("%d seconds", seconds)
	}

	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Account Protection - Too Many Login Attempts</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #e53e3e 0%, #c53030 50%, #9b2c2c 100%);
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
            width: 100%;
            padding: 40px;
            text-align: center;
            animation: slideIn 0.5s ease-out;
        }
        @keyframes slideIn {
            from { opacity: 0; transform: translateY(-20px); }
            to { opacity: 1; transform: translateY(0); }
        }
        .icon {
            width: 100px; height: 100px;
            margin: 0 auto 20px;
            background: linear-gradient(135deg, #e53e3e 0%, #c53030 100%);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            animation: pulse 2s ease-in-out infinite;
        }
        @keyframes pulse {
            0%, 100% { transform: scale(1); }
            50% { transform: scale(1.05); }
        }
        .icon svg { width: 50px; height: 50px; fill: white; }
        h1 { color: #2d3748; font-size: 28px; margin-bottom: 10px; }
        .subtitle { color: #718096; margin-bottom: 25px; font-size: 16px; }
        .lockout-box {
            background: linear-gradient(135deg, #fff5f5 0%, #fed7d7 100%);
            border: 2px solid #fc8181;
            border-radius: 15px;
            padding: 25px;
            margin: 25px 0;
        }
        .lockout-title { color: #c53030; font-weight: 700; font-size: 14px; text-transform: uppercase; letter-spacing: 1px; margin-bottom: 15px; }
        .countdown { font-size: 48px; font-weight: 700; color: #c53030; font-family: 'Courier New', monospace; }
        .countdown-label { color: #718096; font-size: 14px; margin-top: 10px; }
        .info-box {
            background: #f7fafc;
            border-radius: 10px;
            padding: 20px;
            margin: 20px 0;
            text-align: left;
        }
        .info-box h3 { color: #2d3748; font-size: 16px; margin-bottom: 12px; display: flex; align-items: center; }
        .info-box ul { list-style: none; padding: 0; }
        .info-box li { color: #4a5568; font-size: 14px; padding: 8px 0; padding-left: 25px; position: relative; }
        .info-box li:before { content: "→"; position: absolute; left: 0; color: #e53e3e; font-weight: bold; }
        .footer { margin-top: 30px; color: #a0aec0; font-size: 14px; }
        .ip-info { color: #a0aec0; font-size: 12px; margin-top: 10px; }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">
            <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg">
                <path d="M18 8h-1V6c0-2.76-2.24-5-5-5S7 3.24 7 6v2H6c-1.1 0-2 .9-2 2v10c0 1.1.9 2 2 2h12c1.1 0 2-.9 2-2V10c0-1.1-.9-2-2-2zm-6 9c-1.1 0-2-.9-2-2s.9-2 2-2 2 .9 2 2-.9 2-2 2zm3.1-9H8.9V6c0-1.71 1.39-3.1 3.1-3.1 1.71 0 3.1 1.39 3.1 3.1v2z"/>
            </svg>
        </div>
        
        <h1>🔒 Account Protection Active</h1>
        <p class="subtitle">Too many failed login attempts detected</p>
        
        <div class="lockout-box">
            <div class="lockout-title">⏱️ Temporary Lockout</div>
            <div class="countdown" id="countdown">` + timeDisplay + `</div>
            <div class="countdown-label">until you can try again</div>
        </div>
        
        <div class="info-box">
            <h3>💡 Why am I seeing this?</h3>
            <ul>
                <li>Multiple failed login attempts were detected</li>
                <li>This protection prevents unauthorized access</li>
                <li>Wait for the timer or contact support</li>
                <li>If you forgot your password, use password reset</li>
            </ul>
        </div>
        
        <p class="ip-info">Your IP: ` + clientIP + ` | Domain: ` + domain + `</p>
        <p class="footer">🛡️ Protected by DoCode WAF - Brute Force Protection</p>
    </div>
    
    <script>
        let totalSeconds = ` + fmt.Sprintf("%d", lockoutSeconds) + `;
        const countdownEl = document.getElementById('countdown');
        
        function updateCountdown() {
            const minutes = Math.floor(totalSeconds / 60);
            const seconds = totalSeconds % 60;
            
            if (minutes > 0) {
                countdownEl.textContent = minutes + ' min ' + seconds + ' sec';
            } else {
                countdownEl.textContent = seconds + ' seconds';
            }
            
            if (totalSeconds <= 0) {
                countdownEl.textContent = 'Ready!';
                setTimeout(() => window.location.reload(), 1000);
            } else {
                totalSeconds--;
                setTimeout(updateCountdown, 1000);
            }
        }
        
        updateCountdown();
    </script>
</body>
</html>`
}
