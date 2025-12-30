package api

import (
	"fmt"
	"net/http"

	"github.com/aleh/docode-waf/internal/constants"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// BlockingRuleHandler handles blocking rule requests
type BlockingRuleHandler struct {
	db *sqlx.DB
}

// NewBlockingRuleHandler creates a new blocking rule handler
func NewBlockingRuleHandler(db *sqlx.DB) *BlockingRuleHandler {
	return &BlockingRuleHandler{db: db}
}

// ListBlockingRules returns all blocking rules
func (h *BlockingRuleHandler) ListBlockingRules(c *gin.Context) {
	var rules []struct {
		ID        string `db:"id" json:"id"`
		Name      string `db:"name" json:"name"`
		Type      string `db:"type" json:"type"`
		Pattern   string `db:"pattern" json:"pattern"`
		Action    string `db:"action" json:"action"`
		Enabled   bool   `db:"enabled" json:"enabled"`
		Priority  int    `db:"priority" json:"priority"`
		CreatedAt string `db:"created_at" json:"created_at"`
		UpdatedAt string `db:"updated_at" json:"updated_at"`
	}

	query := `
		SELECT id, name, type, pattern, action, enabled, priority, 
		       created_at, updated_at
		FROM blocking_rules 
		ORDER BY priority DESC, created_at DESC
	`

	if err := h.db.Select(&rules, query); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch blocking rules",
		})
		return
	}

	c.JSON(http.StatusOK, rules)
}

// GetBlockingRule returns a specific blocking rule
func (h *BlockingRuleHandler) GetBlockingRule(c *gin.Context) {
	id := c.Param("id")

	var rule struct {
		ID        string `db:"id" json:"id"`
		Name      string `db:"name" json:"name"`
		Type      string `db:"type" json:"type"`
		Pattern   string `db:"pattern" json:"pattern"`
		Action    string `db:"action" json:"action"`
		Enabled   bool   `db:"enabled" json:"enabled"`
		Priority  int    `db:"priority" json:"priority"`
		CreatedAt string `db:"created_at" json:"created_at"`
		UpdatedAt string `db:"updated_at" json:"updated_at"`
	}

	query := `
		SELECT id, name, type, pattern, action, enabled, priority, 
		       created_at, updated_at
		FROM blocking_rules 
		WHERE id = $1
	`

	if err := h.db.Get(&rule, query, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": constants.ErrBlockingRuleNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// CreateBlockingRule creates a new blocking rule
func (h *BlockingRuleHandler) CreateBlockingRule(c *gin.Context) {
	var input struct {
		Name     string `json:"name" binding:"required"`
		Type     string `json:"type" binding:"required,oneof=ip region url user_agent"`
		Pattern  string `json:"pattern" binding:"required"`
		Action   string `json:"action" binding:"required,oneof=block challenge allow"`
		Enabled  bool   `json:"enabled"`
		Priority int    `json:"priority"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	id := uuid.New().String()
	query := `
		INSERT INTO blocking_rules (id, name, type, pattern, action, enabled, priority, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, name, type, pattern, action, enabled, priority, created_at, updated_at
	`

	var rule struct {
		ID        string `db:"id" json:"id"`
		Name      string `db:"name" json:"name"`
		Type      string `db:"type" json:"type"`
		Pattern   string `db:"pattern" json:"pattern"`
		Action    string `db:"action" json:"action"`
		Enabled   bool   `db:"enabled" json:"enabled"`
		Priority  int    `db:"priority" json:"priority"`
		CreatedAt string `db:"created_at" json:"created_at"`
		UpdatedAt string `db:"updated_at" json:"updated_at"`
	}

	err := h.db.QueryRowx(query, id, input.Name, input.Type, input.Pattern, input.Action, input.Enabled, input.Priority).StructScan(&rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create blocking rule",
		})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// UpdateBlockingRule updates an existing blocking rule
func (h *BlockingRuleHandler) UpdateBlockingRule(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Name     string `json:"name"`
		Type     string `json:"type" binding:"omitempty,oneof=ip region url user_agent"`
		Pattern  string `json:"pattern"`
		Action   string `json:"action" binding:"omitempty,oneof=block challenge allow"`
		Enabled  *bool  `json:"enabled"`
		Priority *int   `json:"priority"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Build dynamic update query
	query := `UPDATE blocking_rules SET updated_at = NOW()`
	args := []interface{}{}
	argIndex := 1

	if input.Name != "" {
		query += fmt.Sprintf(`, name = $%d`, argIndex)
		args = append(args, input.Name)
		argIndex++
	}
	if input.Type != "" {
		query += fmt.Sprintf(`, type = $%d`, argIndex)
		args = append(args, input.Type)
		argIndex++
	}
	if input.Pattern != "" {
		query += fmt.Sprintf(`, pattern = $%d`, argIndex)
		args = append(args, input.Pattern)
		argIndex++
	}
	if input.Action != "" {
		query += fmt.Sprintf(`, action = $%d`, argIndex)
		args = append(args, input.Action)
		argIndex++
	}
	if input.Enabled != nil {
		query += fmt.Sprintf(`, enabled = $%d`, argIndex)
		args = append(args, *input.Enabled)
		argIndex++
	}
	if input.Priority != nil {
		query += fmt.Sprintf(`, priority = $%d`, argIndex)
		args = append(args, *input.Priority)
		argIndex++
	}

	query += fmt.Sprintf(` WHERE id = $%d`, argIndex)
	args = append(args, id)

	result, err := h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update blocking rule",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": constants.ErrBlockingRuleNotFound,
		})
		return
	}

	// Return updated rule
	h.GetBlockingRule(c)
}

// DeleteBlockingRule deletes a blocking rule
func (h *BlockingRuleHandler) DeleteBlockingRule(c *gin.Context) {
	id := c.Param("id")

	query := `DELETE FROM blocking_rules WHERE id = $1`
	result, err := h.db.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete blocking rule",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": constants.ErrBlockingRuleNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blocking rule deleted successfully",
	})
}

// ToggleBlockingRule toggles the enabled status of a blocking rule
func (h *BlockingRuleHandler) ToggleBlockingRule(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Enabled bool `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	query := `UPDATE blocking_rules SET enabled = $1, updated_at = NOW() WHERE id = $2`
	result, err := h.db.Exec(query, input.Enabled, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to toggle blocking rule",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": constants.ErrBlockingRuleNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Blocking rule toggled successfully",
		"enabled": input.Enabled,
	})
}
