package middleware

import (
	"net/http"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var badBotPatterns = []string{
	"(?i)(bot|crawler|spider|scraper)",
	"(?i)(wget|curl|python-requests)",
	"(?i)(semrush|ahrefs|majestic)",
}

// BotDetectorMiddleware detects and blocks known bad bots
func BotDetectorMiddleware() gin.HandlerFunc {
	compiledPatterns := make([]*regexp.Regexp, 0, len(badBotPatterns))
	for _, pattern := range badBotPatterns {
		re, err := regexp.Compile(pattern)
		if err == nil {
			compiledPatterns = append(compiledPatterns, re)
		}
	}

	return func(c *gin.Context) {
		userAgent := c.GetHeader("User-Agent")

		// Empty user agent is suspicious
		if userAgent == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Missing User-Agent",
			})
			c.Abort()
			return
		}

		// Check against bad bot patterns
		for _, pattern := range compiledPatterns {
			if pattern.MatchString(userAgent) {
				// Check if it's a legitimate bot (Google, Bing, etc.)
				if !isLegitimateBot(userAgent) {
					c.JSON(http.StatusForbidden, gin.H{
						"error": "Bot access denied",
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

func isLegitimateBot(userAgent string) bool {
	legitimateBots := []string{
		"Googlebot",
		"Bingbot",
		"Slurp", // Yahoo
		"DuckDuckBot",
		"Baiduspider",
		"YandexBot",
		"facebookexternalhit",
		"LinkedInBot",
		"Twitterbot",
	}

	userAgentLower := strings.ToLower(userAgent)
	for _, bot := range legitimateBots {
		if strings.Contains(userAgentLower, strings.ToLower(bot)) {
			return true
		}
	}
	return false
}
