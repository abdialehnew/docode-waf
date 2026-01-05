package api

import (
	"encoding/base64"
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

// decodeID decodes a base64-encoded ID to the original UUID string
func decodeID(encodedID string) (string, error) {
	decoded, err := base64.URLEncoding.DecodeString(encodedID)
	if err != nil {
		// Try standard encoding if URL encoding fails
		decoded, err = base64.StdEncoding.DecodeString(encodedID)
		if err != nil {
			return "", err
		}
	}
	return string(decoded), nil
}

// ListIPGroups returns all IP groups
func (h *IPGroupHandler) ListIPGroups(c *gin.Context) {
	var groups []map[string]interface{}

	query := `
		SELECT ig.id, ig.name, ig.description, ig.type,
		       COALESCE(array_agg(DISTINCT igv.vhost_id) FILTER (WHERE igv.vhost_id IS NOT NULL), '{}') as vhost_ids,
		       COALESCE(array_agg(DISTINCT v.domain) FILTER (WHERE v.domain IS NOT NULL), '{}') as vhost_domains,
		       ig.created_at, ig.updated_at
		FROM ip_groups ig
		LEFT JOIN ip_group_vhosts igv ON ig.id = igv.ip_group_id
		LEFT JOIN vhosts v ON igv.vhost_id = v.id
		GROUP BY ig.id, ig.name, ig.description, ig.type, ig.created_at, ig.updated_at
		ORDER BY ig.created_at DESC
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
	encodedID := c.Param("id")
	id, err := decodeID(encodedID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	group := make(map[string]interface{})
	query := `
		SELECT ig.id, ig.name, ig.description, ig.type,
		       COALESCE(array_agg(DISTINCT igv.vhost_id) FILTER (WHERE igv.vhost_id IS NOT NULL), '{}') as vhost_ids,
		       COALESCE(array_agg(DISTINCT v.domain) FILTER (WHERE v.domain IS NOT NULL), '{}') as vhost_domains,
		       ig.created_at, ig.updated_at
		FROM ip_groups ig
		LEFT JOIN ip_group_vhosts igv ON ig.id = igv.ip_group_id
		LEFT JOIN vhosts v ON igv.vhost_id = v.id
		WHERE ig.id = $1
		GROUP BY ig.id, ig.name, ig.description, ig.type, ig.created_at, ig.updated_at
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
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Type        string   `json:"type" binding:"required"` // whitelist or blacklist
		VhostIDs    []string `json:"vhost_ids"`               // array of vhost IDs, empty for global rules
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx, err := h.db.Beginx()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	// Create IP group
	query := `
		INSERT INTO ip_groups (id, name, description, type, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, $5)
		RETURNING id
	`

	var id string
	err = tx.QueryRow(query,
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

	// Insert vhost associations
	if len(input.VhostIDs) > 0 {
		for _, vhostID := range input.VhostIDs {
			if vhostID != "" {
				_, err = tx.Exec(`
					INSERT INTO ip_group_vhosts (ip_group_id, vhost_id)
					VALUES ($1, $2)
					ON CONFLICT (ip_group_id, vhost_id) DO NOTHING
				`, id, vhostID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "IP Group created successfully"})
}

// UpdateIPGroup updates an existing IP group
func (h *IPGroupHandler) UpdateIPGroup(c *gin.Context) {
	encodedID := c.Param("id")
	id, err := decodeID(encodedID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var input struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Type        string   `json:"type" binding:"required"`
		VhostIDs    []string `json:"vhost_ids"` // array of vhost IDs, empty for global rules
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Start transaction
	tx, err := h.db.Beginx()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer tx.Rollback()

	// Update IP group
	query := `
		UPDATE ip_groups 
		SET name = $1, description = $2, type = $3, updated_at = $4
		WHERE id = $5
	`

	_, err = tx.Exec(query,
		input.Name,
		input.Description,
		input.Type,
		time.Now(),
		id,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Delete existing vhost associations
	_, err = tx.Exec(`DELETE FROM ip_group_vhosts WHERE ip_group_id = $1`, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Insert new vhost associations
	if len(input.VhostIDs) > 0 {
		for _, vhostID := range input.VhostIDs {
			if vhostID != "" {
				_, err = tx.Exec(`
					INSERT INTO ip_group_vhosts (ip_group_id, vhost_id)
					VALUES ($1, $2)
					ON CONFLICT (ip_group_id, vhost_id) DO NOTHING
				`, id, vhostID)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "IP Group updated successfully"})
}

// AddIPAddress adds an IP address to a group
func (h *IPGroupHandler) AddIPAddress(c *gin.Context) {
	encodedGroupID := c.Param("id")
	groupID, err := decodeID(encodedGroupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID format"})
		return
	}

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
	err = h.db.QueryRow(query,
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

// UpdateIPAddress updates an existing IP address
func (h *IPGroupHandler) UpdateIPAddress(c *gin.Context) {
	encodedAddressID := c.Param("addressId")
	addressID, err := decodeID(encodedAddressID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID format"})
		return
	}

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
		UPDATE ip_addresses 
		SET ip_address = $1, cidr_mask = $2, description = $3
		WHERE id = $4
	`

	_, err = h.db.Exec(query,
		input.IPAddress,
		input.CIDRMask,
		input.Description,
		addressID,
	)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "IP Address updated successfully"})
}

// GetIPAddresses returns all IP addresses for a specific group
func (h *IPGroupHandler) GetIPAddresses(c *gin.Context) {
	encodedGroupID := c.Param("id")
	groupID, err := decodeID(encodedGroupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid group ID format"})
		return
	}

	var addresses []map[string]interface{}
	query := `
		SELECT id, ip_address, cidr_mask, description, created_at
		FROM ip_addresses 
		WHERE group_id = $1
		ORDER BY created_at DESC
	`

	rows, err := h.db.Queryx(query, groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		addr := make(map[string]interface{})
		err := rows.MapScan(addr)
		if err != nil {
			continue
		}
		addresses = append(addresses, addr)
	}

	// Return empty array instead of null
	if addresses == nil {
		addresses = []map[string]interface{}{}
	}

	c.JSON(http.StatusOK, addresses)
}

// DeleteIPAddress removes an IP address from a group
func (h *IPGroupHandler) DeleteIPAddress(c *gin.Context) {
	encodedAddressID := c.Param("addressId")
	addressID, err := decodeID(encodedAddressID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid address ID format"})
		return
	}

	_, err = h.db.Exec("DELETE FROM ip_addresses WHERE id = $1", addressID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "IP Address deleted successfully"})
}

// DeleteIPGroup deletes an IP group
func (h *IPGroupHandler) DeleteIPGroup(c *gin.Context) {
	encodedID := c.Param("id")
	id, err := decodeID(encodedID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	// Delete all addresses in the group first
	_, err = h.db.Exec("DELETE FROM ip_addresses WHERE group_id = $1", id)
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
