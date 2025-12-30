package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// IPGroupHandler handles IP group requests
type IPGroupHandler struct {
	db *sqlx.DB
}

// NewIPGroupHandler creates a new IP group handler
func NewIPGroupHandler(db *sqlx.DB) *IPGroupHandler {
	return &IPGroupHandler{db: db}
}

// ListIPGroups returns all IP groups
func (h *IPGroupHandler) ListIPGroups(c *gin.Context) {
	var groups []map[string]interface{}

	query := `
		SELECT id, name, description, type, created_at, updated_at
		FROM ip_groups 
		ORDER BY created_at DESC
	`

	rows, err := h.db.Queryx(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		group := make(map[string]interface{})
		err := rows.MapScan(group)
		if err != nil {
			continue
		}
		groups = append(groups, group)
	}

	c.JSON(http.StatusOK, groups)
}

// GetIPGroup returns a specific IP group with its addresses
func (h *IPGroupHandler) GetIPGroup(c *gin.Context) {
	id := c.Param("id")

	group := make(map[string]interface{})
	query := `
		SELECT id, name, description, type, created_at, updated_at
		FROM ip_groups 
		WHERE id = $1
	`

	rows, err := h.db.Queryx(query, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	if !rows.Next() {
		c.JSON(http.StatusNotFound, gin.H{"error": "IP Group not found"})
		return
	}

	err = rows.MapScan(group)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get IP addresses for this group
	var addresses []map[string]interface{}
	addrQuery := `
		SELECT id, ip_address, cidr_mask, description, created_at
		FROM ip_addresses 
		WHERE group_id = $1
		ORDER BY created_at DESC
	`

	addrRows, err := h.db.Queryx(addrQuery, id)
	if err == nil {
		defer addrRows.Close()
		for addrRows.Next() {
			addr := make(map[string]interface{})
			err := addrRows.MapScan(addr)
			if err == nil {
				addresses = append(addresses, addr)
			}
		}
	}

	group["addresses"] = addresses
	c.JSON(http.StatusOK, group)
}

// CreateIPGroup creates a new IP group
func (h *IPGroupHandler) CreateIPGroup(c *gin.Context) {
	var input struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Type        string `json:"type" binding:"required"` // whitelist or blacklist
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		INSERT INTO ip_groups (id, name, description, type, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
		RETURNING id
	`

	var id string
	err := h.db.QueryRow(query,
		input.Name,
		input.Description,
		input.Type,
		time.Now(),
		time.Now(),
	).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "IP Group created successfully"})
}

// AddIPAddress adds an IP address to a group
func (h *IPGroupHandler) AddIPAddress(c *gin.Context) {
	groupID := c.Param("id")

	var input struct {
		IPAddress   string `json:"ip_address" binding:"required"`
		CIDRMask    *int   `json:"cidr_mask"`
		Description string `json:"description"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query := `
		INSERT INTO ip_addresses (id, group_id, ip_address, cidr_mask, description, created_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
		RETURNING id
	`

	var id string
	err := h.db.QueryRow(query,
		groupID,
		input.IPAddress,
		input.CIDRMask,
		input.Description,
		time.Now(),
	).Scan(&id)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "IP Address added successfully"})
}

// DeleteIPAddress removes an IP address from a group
func (h *IPGroupHandler) DeleteIPAddress(c *gin.Context) {
	addressID := c.Param("addressId")

	_, err := h.db.Exec("DELETE FROM ip_addresses WHERE id = $1", addressID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "IP Address deleted successfully"})
}

// DeleteIPGroup deletes an IP group
func (h *IPGroupHandler) DeleteIPGroup(c *gin.Context) {
	id := c.Param("id")

	// Delete all addresses in the group first
	_, err := h.db.Exec("DELETE FROM ip_addresses WHERE group_id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Delete the group
	_, err = h.db.Exec("DELETE FROM ip_groups WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "IP Group deleted successfully"})
}
