package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// IPBlockerMiddleware blocks requests from blacklisted IPs and allows only whitelisted IPs
func IPBlockerMiddleware(db *sqlx.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// Check whitelist first
		whitelisted, err := isIPInGroup(db, clientIP, "whitelist")
		if err == nil && whitelisted {
			c.Next()
			return
		}

		// Check blacklist
		blacklisted, err := isIPInGroup(db, clientIP, "blacklist")
		if err == nil && blacklisted {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
			})
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
		SELECT COUNT(*) FROM ip_addresses ia
		JOIN ip_groups ig ON ia.group_id = ig.id
		WHERE ig.type = $1 AND ia.ip_address = $2
	`
	var count int
	err := db.Get(&count, query, groupType, clientIP)
	if err != nil {
		return false, err
	}
	return count > 0, nil
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
