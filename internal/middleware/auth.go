package middleware

import (
	"net/http"
	"strings"

	"github.com/aleh/docode-waf/internal/services"
	"github.com/gin-gonic/gin"
)

// AuthMiddleware validates JWT token
func AuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]

		// Validate token
		admin, err := authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set admin info in context
		c.Set("admin_id", admin.ID)
		c.Set("admin", admin)
		c.Set("admin_role", admin.Role)

		c.Next()
	}
}

// OptionalAuthMiddleware validates token if present but doesn't require it
func OptionalAuthMiddleware(authService *services.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				admin, err := authService.ValidateToken(parts[1])
				if err == nil {
					c.Set("admin_id", admin.ID)
					c.Set("admin", admin)
					c.Set("admin_role", admin.Role)
				}
			}
		}
		c.Next()
	}
}

// RoleMiddleware checks if admin has required role
func RoleMiddleware(roles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		adminRole, exists := c.Get("admin_role")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
			c.Abort()
			return
		}

		roleStr := adminRole.(string)
		for _, role := range roles {
			if roleStr == role {
				c.Next()
				return
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
		c.Abort()
	}
}
