package services

import (
	"database/sql"
	"encoding/json"
	"net"
	"regexp"
	"strings"

	"github.com/aleh/docode-waf/internal/models"
	"github.com/gin-gonic/gin"
)

type RuleService struct {
	db *sql.DB
}

func NewRuleService(db *sql.DB) *RuleService {
	return &RuleService{db: db}
}

// GetRules retrieves all enabled rules ordered by priority
func (s *RuleService) GetRules() ([]models.Rule, error) {
	query := `
		SELECT id, name, description, priority, action, match_logic, conditions, enabled, created_at, updated_at
		FROM rules
		ORDER BY priority DESC, created_at DESC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.Rule
	for rows.Next() {
		var r models.Rule
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Priority, &r.Action, &r.MatchLogic, &r.Conditions, &r.Enabled, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}

		// Parse conditions
		if len(r.Conditions) > 0 {
			_ = json.Unmarshal(r.Conditions, &r.ConditionList)
		}

		rules = append(rules, r)
	}
	return rules, nil
}

// GetAllRules retrieves all rules (including disabled) for admin UI
func (s *RuleService) GetAllRules() ([]models.Rule, error) {
	query := `
		SELECT id, name, description, priority, action, match_logic, conditions, enabled, created_at, updated_at
		FROM rules
		ORDER BY priority DESC, created_at DESC
	`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []models.Rule
	for rows.Next() {
		var r models.Rule
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Priority, &r.Action, &r.MatchLogic, &r.Conditions, &r.Enabled, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, err
		}
		rules = append(rules, r)
	}
	return rules, nil
}

func (s *RuleService) CreateRule(r *models.Rule) error {
	query := `
		INSERT INTO rules (name, description, priority, action, match_logic, conditions, enabled)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
	`
	// Validate conditions JSON
	if len(r.Conditions) == 0 {
		r.Conditions = json.RawMessage("[]")
	}

	if r.MatchLogic == "" {
		r.MatchLogic = "AND"
	}

	return s.db.QueryRow(query, r.Name, r.Description, r.Priority, r.Action, r.MatchLogic, r.Conditions, r.Enabled).Scan(&r.ID, &r.CreatedAt, &r.UpdatedAt)
}

func (s *RuleService) UpdateRule(r *models.Rule) error {
	query := `
		UPDATE rules 
		SET name=$2, description=$3, priority=$4, action=$5, match_logic=$6, conditions=$7, enabled=$8, updated_at=CURRENT_TIMESTAMP
		WHERE id=$1
	`
	_, err := s.db.Exec(query, r.ID, r.Name, r.Description, r.Priority, r.Action, r.MatchLogic, r.Conditions, r.Enabled)
	return err
}

func (s *RuleService) DeleteRule(id string) error {
	_, err := s.db.Exec("DELETE FROM rules WHERE id=$1", id)
	return err
}

// EvaluateRequest checks if a request matches any active rules
// Returns: action (string), ruleID (string), matched (bool)
func (s *RuleService) EvaluateRequest(c *gin.Context, rules []models.Rule) (string, string, bool) {
	for _, rule := range rules {
		if !rule.Enabled {
			continue
		}

		if s.matchesConditions(c, rule.ConditionList, rule.MatchLogic) {
			return rule.Action, rule.ID, true
		}
	}
	return "", "", false
}

func (s *RuleService) matchesConditions(c *gin.Context, conditions []models.Condition, logic string) bool {
	if len(conditions) == 0 {
		return false
	}

	// AND logic (default)
	if logic == "" || strings.ToUpper(logic) == "AND" {
		for _, cond := range conditions {
			if !s.checkCondition(c, cond) {
				return false // AND logic: all must match
			}
		}
		return true
	}

	// OR logic
	if strings.ToUpper(logic) == "OR" {
		for _, cond := range conditions {
			if s.checkCondition(c, cond) {
				return true // OR logic: any match is enough
			}
		}
		return false // No match found
	}

	return false
}

func (s *RuleService) checkCondition(c *gin.Context, cond models.Condition) bool {
	var actualValue string

	switch cond.Field {
	case "ip":
		actualValue = c.ClientIP()
	case "method":
		actualValue = c.Request.Method
	case "path":
		actualValue = c.Request.URL.Path
	case "user_agent":
		actualValue = c.Request.UserAgent()
	case "country":
		// This requires GeoIP middleware to set this in context or header
		// For now, let's try to get it from header if set by upstream/logging
		actualValue = c.Writer.Header().Get("X-GeoIP-Country")
		if actualValue == "" {
			// Fallback: This service might need access to GeoIP DB directly ideally,
			// or rely on previous middleware
			actualValue = ""
		}
	case "header":
		// Special case: Value format "HeaderName: ExpectedValue" ???
		// Or we need a mechanism to specify WHICH header.
		// For simplicity in this iteration, let's assume Condition.Field could be "header:User-Agent"
		// But schema says "field": "header".
		// Let's assume the user puts "HeaderName" in operator? No.
		// Let's support specific headers if needed, or basic common ones.
		// For strict field "header", we can't know which one.
		// IMPROVEMENT: Field should be "header.Name"
		actualValue = ""
	default:
		// Support "header.X-Custom"
		if strings.HasPrefix(cond.Field, "header.") {
			headerName := strings.TrimPrefix(cond.Field, "header.")
			actualValue = c.GetHeader(headerName)
		} else if strings.HasPrefix(cond.Field, "query.") {
			queryName := strings.TrimPrefix(cond.Field, "query.")
			actualValue = c.Query(queryName)
		} else {
			return false
		}
	}

	return s.evaluateOperator(actualValue, cond.Operator, cond.Value)
}

func (s *RuleService) evaluateOperator(actual, op, expected string) bool {
	switch op {
	case "eq":
		return actual == expected
	case "neq":
		return actual != expected
	case "contains":
		return strings.Contains(actual, expected)
	case "not_contains":
		return !strings.Contains(actual, expected)
	case "starts_with":
		return strings.HasPrefix(actual, expected)
	case "ends_with":
		return strings.HasSuffix(actual, expected)
	case "regex":
		matched, _ := regexp.MatchString(expected, actual)
		return matched
	case "gt":
		return actual > expected // String comparison
	case "lt":
		return actual < expected
	case "cidr_contains":
		// actual = IP, expected = CIDR
		_, ipNet, err := net.ParseCIDR(expected)
		if err != nil {
			return false
		}
		ip := net.ParseIP(actual)
		return ipNet.Contains(ip)
	}
	return false
}
