package middleware

import (
	"log"
	"net/http"

	"github.com/aleh/docode-waf/internal/services"
	"github.com/gin-gonic/gin"
)

type RuleExecutor struct {
	ruleService *services.RuleService
}

func NewRuleExecutor(service *services.RuleService) *RuleExecutor {
	return &RuleExecutor{ruleService: service}
}

func (e *RuleExecutor) Execute() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Optimization: Check if this is an API call to the dashboard itself?
		// We probably don't want to lock ourselves out of the admin panel easily
		if c.Request.URL.Path == "/api/v1/auth/login" {
			c.Next()
			return
		}

		// Retrieve active rules from DB (cache this in production!)
		// For now, we query every time (not efficient but safe for v1)
		rules, err := e.ruleService.GetRules()
		if err != nil {
			// Fail open on DB error? Or log?
			log.Printf("Error fetching rules: %v", err)
			c.Next()
			return
		}

		action, ruleID, matched := e.ruleService.EvaluateRequest(c, rules)
		if matched {
			log.Printf("[Rule Engine] Match: Rule=%s Action=%s IP=%s Path=%s", ruleID, action, c.ClientIP(), c.Request.URL.Path)

			switch action {
			case "block":
				c.JSON(http.StatusForbidden, gin.H{
					"error":   "Request blocked by custom security rule",
					"rule_id": ruleID,
				})
				c.Abort()
				return
			case "allow":
				// Bypass subsequent WAF checks?
				// For now, we just don't block.
				// To strictly "allow" and skip others, we might need to set a flag in context
				c.Set("WAF_Allow", true)
				c.Next()
				return
			case "log":
				// Already logged above
			case "challenge":
				// Trigger bot challenge (if available) or 403 for now if not integrated
				// Ideally integrate with BotDetector
				// c.Set("ForceChallenge", true)
			}
		}

		c.Next()
	}
}
