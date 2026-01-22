package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// TrafficFiltersHandler handles traffic filter API requests
type TrafficFiltersHandler struct {
	db *sqlx.DB
}

// NewTrafficFiltersHandler creates a new traffic filters handler
func NewTrafficFiltersHandler(db *sqlx.DB) *TrafficFiltersHandler {
	return &TrafficFiltersHandler{db: db}
}

// GetUniqueCountries returns unique country codes from traffic logs for filtering
func (h *TrafficFiltersHandler) GetUniqueCountries(c *gin.Context) {
	var countries []struct {
		CountryCode string `db:"country_code" json:"country_code"`
		Count       int    `db:"count" json:"count"`
	}

	query := `
		SELECT country_code, COUNT(*) as count
		FROM traffic_logs 
		WHERE country_code IS NOT NULL AND country_code != ''
		GROUP BY country_code
		ORDER BY count DESC
	`

	if err := h.db.Select(&countries, query); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": countries,
	})
}

// GetUniqueIPs returns unique IPs from traffic logs, optionally filtered by country
func (h *TrafficFiltersHandler) GetUniqueIPs(c *gin.Context) {
	countryCodes := c.QueryArray("country_code")

	var ips []struct {
		ClientIP    string `db:"client_ip" json:"client_ip"`
		CountryCode string `db:"country_code" json:"country_code"`
		Count       int    `db:"count" json:"count"`
	}

	query := `
		SELECT client_ip, country_code, COUNT(*) as count
		FROM traffic_logs 
		WHERE country_code IS NOT NULL
	`
	args := []interface{}{}

	if len(countryCodes) > 0 {
		query += ` AND country_code = ANY($1)`
		args = append(args, pq.Array(countryCodes))
	}

	query += ` GROUP BY client_ip, country_code ORDER BY count DESC LIMIT 500`

	var err error
	if len(args) > 0 {
		err = h.db.Select(&ips, query, args...)
	} else {
		err = h.db.Select(&ips, query)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": ips,
	})
}
