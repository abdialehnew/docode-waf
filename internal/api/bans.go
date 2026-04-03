package api

import (
	"net/http"
	"time"

	"github.com/aleh/docode-waf/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// BansHandler handles ban management requests
type BansHandler struct {
	db                *sqlx.DB
	dynamicBanService *services.DynamicBanService
}

// NewBansHandler creates a new bans handler
func NewBansHandler(db *sqlx.DB, dynamicBanService *services.DynamicBanService) *BansHandler {
	return &BansHandler{
		db:                db,
		dynamicBanService: dynamicBanService,
	}
}

// ListBans returns all active bans
func (h *BansHandler) ListBans(c *gin.Context) {
	type Ban struct {
		ID             string    `db:"id" json:"id"`
		IPAddress      string    `db:"ip_address" json:"ip_address"`
		Reason         string    `db:"reason" json:"reason"`
		BannedAt       time.Time `db:"banned_at" json:"banned_at"`
		ExpiresAt      time.Time `db:"expires_at" json:"expires_at"`
		ViolationCount int       `db:"violation_count" json:"violation_count"`
	}

	var bans []Ban
	query := `
		SELECT id::text, ip_address, reason, banned_at, expires_at, violation_count
		FROM ip_bans
		WHERE expires_at > NOW()
		ORDER BY banned_at DESC
	`

	err := h.db.Select(&bans, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if bans == nil {
		bans = []Ban{}
	}

	c.JSON(http.StatusOK, bans)
}

// UnbanIP removes a ban for a specific IP
func (h *BansHandler) UnbanIP(c *gin.Context) {
	id := c.Param("id")

	// Get IP from ID first
	var ip string
	err := h.db.Get(&ip, "SELECT ip_address FROM ip_bans WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Ban not found"})
		return
	}

	// Unban via service
	err = h.dynamicBanService.UnbanIP(ip)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unban IP: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "IP unbanned successfully"})
}
