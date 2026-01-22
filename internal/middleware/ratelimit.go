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

// RateLimiterMiddleware implements per-vhost rate limiting using Redis
func RateLimiterMiddleware(redisClient *redis.Client, db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get current vhost domain
		domain := c.Request.Host
		if colonIdx := strings.Index(domain, ":"); colonIdx != -1 {
			domain = domain[:colonIdx]
		}

		// Check if rate limiting is enabled for this vhost
		var vhostSettings struct {
			RateLimitEnabled  bool `db:"rate_limit_enabled"`
			RateLimitRequests int  `db:"rate_limit_requests"`
			RateLimitWindow   int  `db:"rate_limit_window"`
		}

		err := db.Get(&vhostSettings, "SELECT rate_limit_enabled, rate_limit_requests, rate_limit_window FROM vhosts WHERE domain = $1", domain)
		if err != nil {
			log.Printf("[Rate Limiter] Error getting vhost settings for domain %s: %v", domain, err)
			c.Next()
			return
		}

		// If rate limiting is disabled for this vhost, skip
		if !vhostSettings.RateLimitEnabled {
			c.Next()
			return
		}

		clientIP := GetRealClientIP(c)
		key := fmt.Sprintf("ratelimit:%s:%s", domain, clientIP)
		ctx := context.Background()

		// Get current count
		count, err := redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.Next()
			return
		}

		// Check if limit exceeded
		if count >= vhostSettings.RateLimitRequests {
			// Get TTL for reset time
			ttl, _ := redisClient.TTL(ctx, key).Result()
			resetTime := time.Now().Add(ttl).Unix()

			// Set headers
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", vhostSettings.RateLimitRequests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetTime))
			c.Header("Retry-After", fmt.Sprintf("%d", int(ttl.Seconds())))
			c.Header("Content-Type", "text/html; charset=utf-8")

			c.String(http.StatusTooManyRequests, getRateLimitHTML(domain, vhostSettings.RateLimitRequests, vhostSettings.RateLimitWindow, int(ttl.Seconds())))
		}

		// Increment counter
		pipe := redisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Duration(vhostSettings.RateLimitWindow)*time.Second)
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		// Add rate limit headers
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", vhostSettings.RateLimitRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", vhostSettings.RateLimitRequests-count-1))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Duration(vhostSettings.RateLimitWindow)*time.Second).Unix()))

		c.Next()
	}
}

// getRateLimitHTML returns HTML for rate limit exceeded page
func getRateLimitHTML(domain string, limit, window, retryAfter int) string {
	return `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Rate Limit Exceeded - DoCode WAF</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            background: linear-gradient(135deg, #ffa726 0%, #fb8c00 50%, #f57c00 100%);
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
            animation: slideIn 0.5s ease-out;
        }
        
        @keyframes slideIn {
            from {
                opacity: 0;
                transform: translateY(-20px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
        
        .icon {
            width: 100px;
            height: 100px;
            margin: 0 auto 20px;
            background: linear-gradient(135deg, #ffa726 0%, #fb8c00 50%, #f57c00 100%);
            border-radius: 50%;
            display: flex;
            align-items: center;
            justify-content: center;
            animation: pulse 2s ease-in-out infinite;
        }
        
        @keyframes pulse {
            0%, 100% {
                transform: scale(1);
            }
            50% {
                transform: scale(1.05);
            }
        }
        
        .icon svg {
            width: 50px;
            height: 50px;
            fill: white;
        }
        
        h1 {
            color: #2d3748;
            font-size: 32px;
            margin-bottom: 10px;
            font-weight: 700;
        }
        
        .subtitle {
            color: #718096;
            margin-bottom: 30px;
            font-size: 16px;
        }
        
        .info-box {
            background: linear-gradient(135deg, #fff5f5 0%, #fed7d7 100%);
            border: 2px solid #fc8181;
            border-radius: 15px;
            padding: 25px;
            margin: 25px 0;
        }
        
        .info-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 12px 0;
            border-bottom: 1px solid rgba(252, 129, 129, 0.3);
        }
        
        .info-item:last-child {
            border-bottom: none;
        }
        
        .info-label {
            color: #c53030;
            font-weight: 600;
            font-size: 14px;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }
        
        .info-value {
            color: #2d3748;
            font-weight: 700;
            font-size: 18px;
        }
        
        .countdown-box {
            background: linear-gradient(135deg, #e6fffa 0%, #b2f5ea 100%);
            border: 2px solid #4fd1c5;
            border-radius: 15px;
            padding: 20px;
            margin: 20px 0;
        }
        
        .countdown-label {
            color: #234e52;
            font-size: 14px;
            font-weight: 600;
            margin-bottom: 10px;
        }
        
        .countdown-timer {
            font-size: 48px;
            font-weight: 700;
            color: #2c7a7b;
            font-family: 'Courier New', monospace;
            letter-spacing: 2px;
        }
        
        .countdown-unit {
            font-size: 16px;
            color: #319795;
            margin-left: 5px;
        }
        
        .progress-bar {
            width: 100%;
            height: 8px;
            background: #e6fffa;
            border-radius: 10px;
            overflow: hidden;
            margin-top: 15px;
        }
        
        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #4fd1c5 0%, #38b2ac 100%);
            border-radius: 10px;
            transition: width 1s linear;
        }
        
        .instructions {
            background: #f7fafc;
            border-radius: 10px;
            padding: 20px;
            margin: 20px 0;
            text-align: left;
        }
        
        .instructions h3 {
            color: #2d3748;
            font-size: 16px;
            margin-bottom: 12px;
            display: flex;
            align-items: center;
        }
        
        .instructions ul {
            list-style: none;
            padding: 0;
        }
        
        .instructions li {
            color: #4a5568;
            font-size: 14px;
            padding: 8px 0;
            padding-left: 25px;
            position: relative;
        }
        
        .instructions li:before {
            content: "→";
            position: absolute;
            left: 0;
            color: #4fd1c5;
            font-weight: bold;
        }
        
        .footer {
            margin-top: 30px;
            color: #a0aec0;
            font-size: 14px;
        }
        
        .domain {
            color: #667eea;
            font-weight: 600;
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
        
        <h1>⚠️ Rate Limit Exceeded</h1>
        <p class="subtitle">You've made too many requests. Please wait before trying again.</p>
        
        <div class="info-box">
            <div class="info-item">
                <span class="info-label">🌐 Domain</span>
                <span class="info-value domain">` + domain + `</span>
            </div>
            <div class="info-item">
                <span class="info-label">📊 Request Limit</span>
                <span class="info-value">` + fmt.Sprintf("%d", limit) + ` requests</span>
            </div>
            <div class="info-item">
                <span class="info-label">⏱️ Time Window</span>
                <span class="info-value">` + fmt.Sprintf("%d", window) + ` seconds</span>
            </div>
        </div>
        
        <div class="countdown-box">
            <div class="countdown-label">⏳ Retry Available In</div>
            <div class="countdown-timer">
                <span id="countdown">` + fmt.Sprintf("%d", retryAfter) + `</span>
                <span class="countdown-unit">sec</span>
            </div>
            <div class="progress-bar">
                <div class="progress-fill" id="progress"></div>
            </div>
        </div>
        
        <div class="instructions">
            <h3>💡 What can you do?</h3>
            <ul>
                <li>Wait for the countdown to finish</li>
                <li>The page will automatically reload when ready</li>
                <li>Reduce the frequency of your requests</li>
                <li>Contact support if you need higher limits</li>
            </ul>
        </div>
        
        <p class="footer">
            🛡️ Protected by DoCode WAF
        </p>
    </div>
    
    <script>
        let secondsLeft = ` + fmt.Sprintf("%d", retryAfter) + `;
        const totalSeconds = secondsLeft;
        const countdownEl = document.getElementById('countdown');
        const progressEl = document.getElementById('progress');
        
        // Set initial progress
        progressEl.style.width = '100%';
        
        const timer = setInterval(() => {
            secondsLeft--;
            countdownEl.textContent = secondsLeft;
            
            // Update progress bar
            const percentage = (secondsLeft / totalSeconds) * 100;
            progressEl.style.width = percentage + '%';
            
            if (secondsLeft <= 0) {
                clearInterval(timer);
                countdownEl.textContent = '0';
                progressEl.style.width = '0%';
                
                // Show reload message
                document.querySelector('.countdown-label').textContent = '✅ Ready! Reloading...';
                
                // Reload after a short delay
                setTimeout(() => {
                    window.location.reload();
                }, 1000);
            }
        }, 1000);
    </script>
</body>
</html>`
}
