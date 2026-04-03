package models

import (
	"encoding/json"
	"time"
)

// Condition represents a single logic check in a rule
// e.g. "Path starts_with /admin"
type Condition struct {
	Field    string `json:"field"`    // ip, country, user_agent, path, method, header, query_param
	Operator string `json:"operator"` // eq, neq, contains, not_contains, starts_with, ends_with, regex, gt, lt
	Value    string `json:"value"`    // The value to compare against
}

// Rule represents a user-defined WAF rule
type Rule struct {
	ID          string          `db:"id" json:"id"`
	Name        string          `db:"name" json:"name"`
	Description string          `db:"description" json:"description"`
	Priority    int             `db:"priority" json:"priority"`
	Action      string          `db:"action" json:"action"`           // block, allow, challenge, log
	MatchLogic  string          `db:"match_logic" json:"match_logic"` // AND, OR
	Conditions  json.RawMessage `db:"conditions" json:"conditions"`
	Enabled     bool            `db:"enabled" json:"enabled"`
	CreatedAt   time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time       `db:"updated_at" json:"updated_at"`

	// Parsed conditions (helper)
	ConditionList []Condition `json:"-"`
}
