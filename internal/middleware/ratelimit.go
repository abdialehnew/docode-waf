package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiterMiddleware implements rate limiting using Redis
func RateLimiterMiddleware(redisClient *redis.Client, requestsPerSecond int, burst int) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		key := fmt.Sprintf("ratelimit:%s", clientIP)
		ctx := context.Background()

		// Get current count
		count, err := redisClient.Get(ctx, key).Int()
		if err != nil && err != redis.Nil {
			c.Next()
			return
		}

		// Check if limit exceeded
		if count >= burst {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
			})
			c.Abort()
			return
		}

		// Increment counter
		pipe := redisClient.Pipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, time.Second)
		_, err = pipe.Exec(ctx)
		if err != nil {
			c.Next()
			return
		}

		c.Next()
	}
}
