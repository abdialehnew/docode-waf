package api

import (
	"database/sql"
	"net/http"
	"github.com/aleh/docode-waf/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)
type IPGroupHandler struct {
	db *sqlx.DB
}
func NewIPGroupHandler(db *sqlx.DB) *IPGroupHandler {
	return &IPGroupHandler{db: db}
func (h *IPGroupHandler) List(c *gin.Context) {
	var groups []models.IPGroup
	err := h.db.Select(&groups, "SELECT * FROM ip_groups ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
func (h *IPGroupHandler) Get(c *gin.Context) {
	id := c.Param("id")
	var group models.IPGroup
	err := h.db.Get(&group, "SELECT * FROM ip_groups WHERE id = $1", id)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "IP Group not found"})
}	c.JSON(http.StatusOK, gin.H{"message": "IP removed successfully"})	}		return		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})	if err != nil {	_, err := h.db.Exec("DELETE FROM ip_addresses WHERE id = $1", ipID)	ipID := c.Param("ipId")func (h *IPGroupHandler) RemoveIP(c *gin.Context) {}	c.JSON(http.StatusOK, ips)	}		return		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})	if err != nil {	err := h.db.Select(&ips, "SELECT * FROM ip_addresses WHERE group_id = $1", groupID)	var ips []models.IPAddress	groupID := c.Param("id")func (h *IPGroupHandler) ListIPs(c *gin.Context) {}	c.JSON(http.StatusCreated, ip)	}		return		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})	if err != nil {		Scan(&ip.ID, &ip.CreatedAt)	err := h.db.QueryRowx(query, ip.GroupID, ip.IPAddress, ip.CIDRMask, ip.Description).				  VALUES ($1, $2, $3, $4) RETURNING id, created_at`	query := `INSERT INTO ip_addresses (group_id, ip_address, cidr_mask, description)	ip.GroupID = groupID	}		return		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})	if err := c.ShouldBindJSON(&ip); err != nil {	var ip models.IPAddress	groupID := c.Param("id")func (h *IPGroupHandler) AddIP(c *gin.Context) {}	c.JSON(http.StatusOK, gin.H{"message": "IP Group deleted successfully"})	}		return		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})	if err != nil {	_, err := h.db.Exec("DELETE FROM ip_groups WHERE id = $1", id)	id := c.Param("id")func (h *IPGroupHandler) Delete(c *gin.Context) {}	c.JSON(http.StatusCreated, group)	}		return		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})	if err != nil {		Scan(&group.ID, &group.CreatedAt, &group.UpdatedAt)	err := h.db.QueryRowx(query, group.Name, group.Description, group.Type).				  VALUES ($1, $2, $3) RETURNING id, created_at, updated_at`	query := `INSERT INTO ip_groups (name, description, type)	}		return		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})	if err := c.ShouldBindJSON(&group); err != nil {	var group models.IPGroupfunc (h *IPGroupHandler) Create(c *gin.Context) {}	c.JSON(http.StatusOK, group)	}		return		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})		}			return
