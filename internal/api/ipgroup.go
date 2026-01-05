package api

import (
	"encoding/base64"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
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
// If the ID is already a valid UUID format, return it as-is
func decodeID(encodedID string) (string, error) {
	// Check if it's already a valid UUID format (with hyphens)
	// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
	if len(encodedID) == 36 && encodedID[8] == '-' && encodedID[13] == '-' && encodedID[18] == '-' && encodedID[23] == '-' {
		return encodedID, nil
	}

	// Try to decode as base64
	decoded, err := base64.URLEncoding.DecodeString(encodedID)
	if err != nil {
		// Try standard encoding if URL encoding fails
		decoded, err = base64.StdEncoding.DecodeString(encodedID)
		if err != nil {
			// If both fail, return the original string (might be a plain UUID)
			return encodedID, nil
		}
	}
	return string(decoded), nil
}

// ListIPGroups returns all IP groups
func (h *IPGroupHandler) ListIPGroups(c *gin.Context) {
	type IPGroupResult struct {
		ID           string         `db:"id"`
		Name         string         `db:"name"`
		Description  string         `db:"description"`
		Type         string         `db:"type"`
		VhostIDs     pq.StringArray `db:"vhost_ids"`
		VhostDomains pq.StringArray `db:"vhost_domains"`
		CreatedAt    time.Time      `db:"created_at"`
		UpdatedAt    time.Time      `db:"updated_at"`
	}

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

	var results []IPGroupResult
	err := h.db.Select(&results, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Convert to response format
	groups := make([]map[string]interface{}, 0, len(results))
	for _, r := range results {
		groups = append(groups, map[string]interface{}{
			"id":            r.ID,
			"name":          r.Name,
			"description":   r.Description,
			"type":          r.Type,
			"vhost_ids":     []string(r.VhostIDs),
			"vhost_domains": []string(r.VhostDomains),
			"created_at":    r.CreatedAt,
			"updated_at":    r.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, groups)
}

// GetIPGroup returns a specific IP group with its addresses
func (h *IPGroupHandler) GetIPGroup(c *gin.Context) {
	type IPGroupResult struct {
		ID           string         `db:"id"`
		Name         string         `db:"name"`
		Description  string         `db:"description"`
		Type         string         `db:"type"`
		VhostIDs     pq.StringArray `db:"vhost_ids"`
		VhostDomains pq.StringArray `db:"vhost_domains"`
		CreatedAt    time.Time      `db:"created_at"`
		UpdatedAt    time.Time      `db:"updated_at"`
	}

	encodedID := c.Param("id")
	id, err := decodeID(encodedID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var result IPGroupResult
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

	err = h.db.Get(&result, query, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "IP Group not found"})
		return
	}

	// Convert to response format
	group := map[string]interface{}{
		"id":            result.ID,
		"name":          result.Name,
		"description":   result.Description,
		"type":          result.Type,
		"vhost_ids":     []string(result.VhostIDs),
		"vhost_domains": []string(result.VhostDomains),
		"created_at":    result.CreatedAt,
		"updated_at":    result.UpdatedAt,
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

	result, err := tx.Exec(query,
		input.Name,
		input.Description,
		input.Type,
		time.Now(),
		id,
	)

	if err != nil {
		log.Printf("[IPGroup] Failed to update group %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update IP group: " + err.Error()})
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		log.Printf("[IPGroup] Group %s not found", id)
		c.JSON(http.StatusNotFound, gin.H{"error": "IP Group not found"})
		return
	}

	// Delete existing vhost associations
	_, err = tx.Exec(`DELETE FROM ip_group_vhosts WHERE ip_group_id = $1`, id)
	if err != nil {
		log.Printf("[IPGroup] Failed to delete vhost associations for group %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete vhost associations: " + err.Error()})
		return
	}

	// Insert new vhost associations
	if len(input.VhostIDs) > 0 {
		log.Printf("[IPGroup] Inserting %d vhost associations for group %s", len(input.VhostIDs), id)
		for _, vhostID := range input.VhostIDs {
			if vhostID != "" {
				_, insertErr := tx.Exec(`
					INSERT INTO ip_group_vhosts (ip_group_id, vhost_id)
					VALUES ($1, $2)
					ON CONFLICT (ip_group_id, vhost_id) DO NOTHING
				`, id, vhostID)
				if insertErr != nil {
					log.Printf("[IPGroup] Failed to insert vhost association %s for group %s: %v", vhostID, id, insertErr)
					c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to insert vhost association: " + insertErr.Error()})
					return
				}
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		log.Printf("[IPGroup] Failed to commit transaction for group %s: %v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit changes: " + err.Error()})
		return
	}

	log.Printf("[IPGroup] Successfully updated group %s", id)
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
