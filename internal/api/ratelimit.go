package api

import (
	"fmt"
	"net/http"

	"github.com/aleh/docode-waf/internal/constants"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// RateLimitHandler handles rate limit rule requests
type RateLimitHandler struct {
	db *sqlx.DB
}

// NewRateLimitHandler creates a new rate limit handler
func NewRateLimitHandler(db *sqlx.DB) *RateLimitHandler {
	return &RateLimitHandler{db: db}
}

// ListRateLimitRules returns all rate limit rules
func (h *RateLimitHandler) ListRateLimitRules(c *gin.Context) {
	var rules []struct {
		ID                string `db:"id" json:"id"`
		Name              string `db:"name" json:"name"`
		PathPattern       string `db:"path_pattern" json:"path_pattern"`
		RequestsPerSecond int    `db:"requests_per_second" json:"requests_per_second"`
		Burst             int    `db:"burst" json:"burst"`
		Enabled           bool   `db:"enabled" json:"enabled"`
		CreatedAt         string `db:"created_at" json:"created_at"`
		UpdatedAt         string `db:"updated_at" json:"updated_at"`
	}

	query := `
		SELECT id, name, path_pattern, requests_per_second, burst, enabled, 
		       created_at, updated_at
		FROM rate_limit_rules 
		ORDER BY created_at DESC
	`

	if err := h.db.Select(&rules, query); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch rate limit rules",
		})
		return
	}

	c.JSON(http.StatusOK, rules)
}

// GetRateLimitRule returns a specific rate limit rule
func (h *RateLimitHandler) GetRateLimitRule(c *gin.Context) {
	id := c.Param("id")

	var rule struct {
		ID                string `db:"id" json:"id"`
		Name              string `db:"name" json:"name"`
		PathPattern       string `db:"path_pattern" json:"path_pattern"`
		RequestsPerSecond int    `db:"requests_per_second" json:"requests_per_second"`
		Burst             int    `db:"burst" json:"burst"`
		Enabled           bool   `db:"enabled" json:"enabled"`
		CreatedAt         string `db:"created_at" json:"created_at"`
		UpdatedAt         string `db:"updated_at" json:"updated_at"`
	}

	query := `
		SELECT id, name, path_pattern, requests_per_second, burst, enabled, 
		       created_at, updated_at
		FROM rate_limit_rules 
		WHERE id = $1
	`

	if err := h.db.Get(&rule, query, id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": constants.ErrRateLimitRuleNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, rule)
}

// CreateRateLimitRule creates a new rate limit rule
func (h *RateLimitHandler) CreateRateLimitRule(c *gin.Context) {
	var input struct {
		Name              string `json:"name" binding:"required"`
		PathPattern       string `json:"path_pattern" binding:"required"`
		RequestsPerSecond int    `json:"requests_per_second" binding:"required,min=1"`
		Burst             int    `json:"burst" binding:"required,min=1"`
		Enabled           bool   `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	id := uuid.New().String()
	query := `
		INSERT INTO rate_limit_rules (id, name, path_pattern, requests_per_second, burst, enabled, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id, name, path_pattern, requests_per_second, burst, enabled, created_at, updated_at
	`

	var rule struct {
		ID                string `db:"id" json:"id"`
		Name              string `db:"name" json:"name"`
		PathPattern       string `db:"path_pattern" json:"path_pattern"`
		RequestsPerSecond int    `db:"requests_per_second" json:"requests_per_second"`
		Burst             int    `db:"burst" json:"burst"`
		Enabled           bool   `db:"enabled" json:"enabled"`
		CreatedAt         string `db:"created_at" json:"created_at"`
		UpdatedAt         string `db:"updated_at" json:"updated_at"`
	}

	err := h.db.QueryRowx(query, id, input.Name, input.PathPattern, input.RequestsPerSecond, input.Burst, input.Enabled).StructScan(&rule)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create rate limit rule",
		})
		return
	}

	c.JSON(http.StatusCreated, rule)
}

// UpdateRateLimitRule updates an existing rate limit rule
func (h *RateLimitHandler) UpdateRateLimitRule(c *gin.Context) {
	id := c.Param("id")

	var input struct {
		Name              string `json:"name"`
		PathPattern       string `json:"path_pattern"`
		RequestsPerSecond *int   `json:"requests_per_second" binding:"omitempty,min=1"`
		Burst             *int   `json:"burst" binding:"omitempty,min=1"`
		Enabled           *bool  `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	// Build dynamic update query
	query := `UPDATE rate_limit_rules SET updated_at = NOW()`
	args := []interface{}{}
	argIndex := 1

	if input.Name != "" {
		query += fmt.Sprintf(`, name = $%d`, argIndex)
		args = append(args, input.Name)
		argIndex++
	}
	if input.PathPattern != "" {
		query += fmt.Sprintf(`, path_pattern = $%d`, argIndex)
		args = append(args, input.PathPattern)
		argIndex++
	}
	if input.RequestsPerSecond != nil {
		query += fmt.Sprintf(`, requests_per_second = $%d`, argIndex)
		args = append(args, *input.RequestsPerSecond)
		argIndex++
	}
	if input.Burst != nil {
		query += fmt.Sprintf(`, burst = $%d`, argIndex)
		args = append(args, *input.Burst)
		argIndex++
	}
	if input.Enabled != nil {
		query += fmt.Sprintf(`, enabled = $%d`, argIndex)
		args = append(args, *input.Enabled)
		argIndex++
	}

	query += fmt.Sprintf(` WHERE id = $%d`, argIndex)
	args = append(args, id)

	result, err := h.db.Exec(query, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update rate limit rule",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": constants.ErrRateLimitRuleNotFound,
		})
		return
	}

	// Return updated rule
	h.GetRateLimitRule(c)
}

// DeleteRateLimitRule deletes a rate limit rule
func (h *RateLimitHandler) DeleteRateLimitRule(c *gin.Context) {
	id := c.Param("id")

	query := `DELETE FROM rate_limit_rules WHERE id = $1`
	result, err := h.db.Exec(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete rate limit rule",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": constants.ErrRateLimitRuleNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rate limit rule deleted successfully",
	})
}

// ToggleRateLimitRule toggles the enabled status of a rate limit rule
func (h *RateLimitHandler) ToggleRateLimitRule(c *gin.Context) {
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

	query := `UPDATE rate_limit_rules SET enabled = $1, updated_at = NOW() WHERE id = $2`
	result, err := h.db.Exec(query, input.Enabled, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to toggle rate limit rule",
		})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error": constants.ErrRateLimitRuleNotFound,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rate limit rule toggled successfully",
		"enabled": input.Enabled,
	})
}
