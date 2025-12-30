package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// HTTPFloodProtectionMiddleware protects against HTTP flood attacks
func HTTPFloodProtectionMiddleware(redisClient *redis.Client, maxRequests int, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("httpflood:%s", clientIP)
		ctx := context.Background()

		// Get request count in time window
		count, err := redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.Next()
			return
		}

		// Check if threshold exceeded
		if count >= maxRequests {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests detected",
			})
			c.Abort()
			return
		}

		// Increment counter with expiration
		pipe := redisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, window)
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		c.Next()
	}
}
