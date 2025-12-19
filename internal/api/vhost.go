package api

import (
	"database/sql"
	"net/http"
	"github.com/aleh/docode-waf/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)
type VHostHandler struct {
	db *sqlx.DB
}
func NewVHostHandler(db *sqlx.DB) *VHostHandler {
	return &VHostHandler{db: db}
func (h *VHostHandler) List(c *gin.Context) {
	var vhosts []models.VHost
	err := h.db.Select(&vhosts, "SELECT * FROM vhosts ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, vhosts)
}	c.JSON(http.StatusOK, gin.H{"message": "VHost deleted successfully"})	}		return		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})	if err != nil {	_, err := h.db.Exec("DELETE FROM vhosts WHERE id = $1", id)	id := c.Param("id")func (h *VHostHandler) Delete(c *gin.Context) {}	c.JSON(http.StatusOK, gin.H{"message": "VHost updated successfully"})	}		return		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})	if err != nil {	)		id,		vhost.Enabled,		vhost.SSLKeyPath,		vhost.SSLCertPath,		vhost.SSLEnabled,		vhost.BackendURL,		vhost.Domain,		vhost.Name,	_, err := h.db.Exec(query,				  WHERE id=$8`			  SET name=$1, domain=$2, backend_url=$3, ssl_enabled=$4, ssl_cert_path=$5, ssl_key_path=$6, enabled=$7	query := `UPDATE vhosts 	}		return		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})	if err := c.ShouldBindJSON(&vhost); err != nil {	var vhost models.VHost	id := c.Param("id")func (h *VHostHandler) Update(c *gin.Context) {}	c.JSON(http.StatusCreated, vhost)	}		return		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})	if err != nil {	).Scan(&vhost.ID, &vhost.CreatedAt, &vhost.UpdatedAt)		vhost.Enabled,		vhost.SSLKeyPath,		vhost.SSLCertPath,		vhost.SSLEnabled,		vhost.BackendURL,		vhost.Domain,		vhost.Name,	err := h.db.QueryRowx(query,				  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`	query := `INSERT INTO vhosts (name, domain, backend_url, ssl_enabled, ssl_cert_path, ssl_key_path, enabled)	}		return		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})	if err := c.ShouldBindJSON(&vhost); err != nil {	var vhost models.VHostfunc (h *VHostHandler) Create(c *gin.Context) {}	c.JSON(http.StatusOK, vhost)	}		return		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})		}			return			c.JSON(http.StatusNotFound, gin.H{"error": "VHost not found"})		if err == sql.ErrNoRows {	if err != nil {	err := h.db.Get(&vhost, "SELECT * FROM vhosts WHERE id = $1", id)	var vhost models.VHost	id := c.Param("id")func (h *VHostHandler) Get(c *gin.Context) {
