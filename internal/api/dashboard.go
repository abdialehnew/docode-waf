package api

import (
	"net/http"

	"github.com/aleh/docode-waf/internal/constants"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

const (
	timeStartSuffix     = " 00:00:00' AND timestamp <= '"
	timeEndSuffix       = " 23:59:59'"
	interval24Hours     = "24 hours"
	intervalQueryPrefix = " AND timestamp >= NOW() - INTERVAL '"
	intervalWherePrefix = " WHERE timestamp >= NOW() - INTERVAL '"
)

// DashboardHandler handles dashboard requests
type DashboardHandler struct {
	db *sqlx.DB
}

// NewDashboardHandler creates a new dashboard handler
func NewDashboardHandler(db *sqlx.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

// GetStats returns dashboard statistics
func (h *DashboardHandler) GetStats(c *gin.Context) {
	var totalRequests, blockedRequests, activeVHosts, attackCount, uniqueIPs int

	// Parse date range from query params
	startDate := c.Query("start")
	endDate := c.Query("end")
	rangeParam := c.Query("range")

	// Build WHERE clause for date filtering
	whereClause := ""
	andClause := ""

	if startDate != "" && endDate != "" {
		// Custom date range
		whereClause = " WHERE timestamp >= '" + startDate + timeStartSuffix + endDate + timeEndSuffix
		andClause = " AND timestamp >= '" + startDate + timeStartSuffix + endDate + timeEndSuffix
	} else if rangeParam != "" {
		// Predefined ranges
		var interval string
		switch rangeParam {
		case "1h":
			interval = "1 hour"
		case "24h":
			interval = interval24Hours
		case "7d":
			interval = "7 days"
		case "30d":
			interval = "30 days"
		default:
			interval = interval24Hours
		}
		whereClause = intervalWherePrefix + interval + "'"
		andClause = intervalQueryPrefix + interval + "'"
	} else {
		// Default to last 24 hours if no range specified
		whereClause = intervalWherePrefix + interval24Hours + "'"
		andClause = intervalQueryPrefix + interval24Hours + "'"
	}

	// Get total requests
	h.db.Get(&totalRequests, constants.SQLCountTrafficLogs+whereClause)

	// Get blocked requests
	h.db.Get(&blockedRequests, constants.SQLCountTrafficLogs+whereClause+" AND blocked = true")

	// Get active vhosts
	h.db.Get(&activeVHosts, constants.SQLCountVHosts)

	// Get attack count from traffic_logs
	h.db.Get(&attackCount, constants.SQLCountTrafficLogs+whereClause+" AND is_attack = true")

	// Get unique IPs
	h.db.Get(&uniqueIPs, constants.SQLCountDistinctIPs+whereClause)

	// Get top attack types with date filter
	var topAttackTypes []map[string]interface{}
	topAttackQuery := `
		SELECT 
			attack_type as name,
			COUNT(*) as value
		FROM traffic_logs 
		WHERE is_attack = true AND attack_type IS NOT NULL` +
		andClause + `
		GROUP BY attack_type
		ORDER BY value DESC
		LIMIT 10
	`
	rows, _ := h.db.Queryx(topAttackQuery)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			stat := make(map[string]interface{})
			rows.MapScan(stat)
			topAttackTypes = append(topAttackTypes, stat)
		}
	}

	// Get recent attacks with date filter
	var recentAttacks []map[string]interface{}
	recentAttacksQuery := `
		SELECT 
			timestamp,
			client_ip,
			attack_type,
			url,
			blocked,
			country_code,
			host,
			CASE 
				WHEN blocked = true THEN 'high'
				ELSE 'medium'
			END as severity
		FROM traffic_logs 
		WHERE is_attack = true` +
		andClause + `
		ORDER BY timestamp DESC 
		LIMIT 20
	`
	rows2, _ := h.db.Queryx(recentAttacksQuery)
	if rows2 != nil {
		defer rows2.Close()
		for rows2.Next() {
			attack := make(map[string]interface{})
			rows2.MapScan(attack)
			recentAttacks = append(recentAttacks, attack)
		}
	}

	// Get traffic by hour with date filter
	var requestsByHour []map[string]interface{}
	timeRangeQuery := `
		SELECT 
			DATE_TRUNC('hour', timestamp) as hour,
			COUNT(*) as count,
			SUM(CASE WHEN blocked = true THEN 1 ELSE 0 END) as blocked
		FROM traffic_logs` +
		whereClause + `
		GROUP BY DATE_TRUNC('hour', timestamp)
		ORDER BY hour ASC
	`
	rows3, _ := h.db.Queryx(timeRangeQuery)
	if rows3 != nil {
		defer rows3.Close()
		for rows3.Next() {
			stat := make(map[string]interface{})
			rows3.MapScan(stat)
			requestsByHour = append(requestsByHour, stat)
		}
	}

	stats := gin.H{
		"total_requests":   totalRequests,
		"blocked_requests": blockedRequests,
		"active_vhosts":    activeVHosts,
		"attack_count":     attackCount,
		"total_attacks":    attackCount, // Alias for frontend compatibility
		"unique_ips":       uniqueIPs,
		"top_attack_types": topAttackTypes,
		"recent_attacks":   recentAttacks,
		"requests_by_hour": requestsByHour,
	}

	c.JSON(http.StatusOK, stats)
}

// GetTrafficLogs returns recent traffic logs
func (h *DashboardHandler) GetTrafficLogs(c *gin.Context) {
	var logs []map[string]interface{}

	// Get limit and offset from query params
	limit := c.DefaultQuery("limit", "100")
	offset := c.DefaultQuery("offset", "0")

	query := `
		SELECT id, timestamp, client_ip, method, url, status_code, 
		       response_time, user_agent, blocked, block_reason, country_code, is_attack, attack_type, host
		FROM traffic_logs 
		ORDER BY timestamp DESC 
		LIMIT $1 OFFSET $2
	`

	rows, err := h.db.Queryx(query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		log := make(map[string]interface{})
		err := rows.MapScan(log)
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}

	c.JSON(http.StatusOK, logs)
}

// GetAttackLogs returns recent attack logs
func (h *DashboardHandler) GetAttackLogs(c *gin.Context) {
	var logs []map[string]interface{}

	query := `
		SELECT id, timestamp, client_ip, method, url, status_code,
		       attack_type, blocked, block_reason, user_agent
		FROM traffic_logs 
		WHERE is_attack = true
		ORDER BY timestamp DESC 
		LIMIT 100
	`

	rows, err := h.db.Queryx(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		log := make(map[string]interface{})
		err := rows.MapScan(log)
		if err != nil {
			continue
		}
		logs = append(logs, log)
	}

	c.JSON(http.StatusOK, logs)
}

// GetAttackStats returns attack statistics grouped by type
func (h *DashboardHandler) GetAttackStats(c *gin.Context) {
	var stats []map[string]interface{}

	query := `
		SELECT 
			attack_type,
			COUNT(*) as count,
			COUNT(CASE WHEN blocked = true THEN 1 END) as blocked_count
		FROM traffic_logs 
		WHERE is_attack = true AND attack_type IS NOT NULL
		GROUP BY attack_type
		ORDER BY count DESC
		LIMIT 10
	`

	rows, err := h.db.Queryx(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		stat := make(map[string]interface{})
		err := rows.MapScan(stat)
		if err != nil {
			continue
		}
		stats = append(stats, stat)
	}

	c.JSON(http.StatusOK, stats)
}

// GetRecentAttacks returns recent attacks with details
func (h *DashboardHandler) GetRecentAttacks(c *gin.Context) {
	var attacks []map[string]interface{}

	query := `
		SELECT 
			timestamp,
			client_ip,
			attack_type,
			url,
			blocked,
			user_agent
		FROM traffic_logs 
		WHERE is_attack = true
		ORDER BY timestamp DESC 
		LIMIT 20
	`

	rows, err := h.db.Queryx(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		attack := make(map[string]interface{})
		err := rows.MapScan(attack)
		if err != nil {
			continue
		}
		attacks = append(attacks, attack)
	}

	c.JSON(http.StatusOK, attacks)
}

// GetAttacksByCountry returns attack statistics by country
func (h *DashboardHandler) GetAttacksByCountry(c *gin.Context) {
	var stats []map[string]interface{}

	// Parse date range from query params
	startDate := c.Query("start")
	endDate := c.Query("end")
	rangeParam := c.Query("range")

	// Build WHERE clause for date filtering
	whereClause := "WHERE is_attack = true"

	if startDate != "" && endDate != "" {
		whereClause += " AND timestamp >= '" + startDate + " 00:00:00' AND timestamp <= '" + endDate + " 23:59:59'"
	} else if rangeParam != "" {
		var interval string
		switch rangeParam {
		case "1h":
			interval = "1 hour"
		case "24h":
			interval = "24 hours"
		case "7d":
			interval = "7 days"
		case "30d":
			interval = "30 days"
		default:
			interval = "24 hours"
		}
		whereClause += " AND timestamp >= NOW() - INTERVAL '" + interval + "'"
	}

	query := `
		SELECT 
			country_code,
			COUNT(*) as total_attacks,
			COUNT(CASE WHEN blocked = true THEN 1 END) as blocked_attacks,
			COUNT(DISTINCT client_ip) as unique_ips,
			COUNT(DISTINCT attack_type) as attack_types
		FROM traffic_logs 
		` + whereClause + `
		  AND country_code IS NOT NULL
		GROUP BY country_code
		ORDER BY total_attacks DESC
		LIMIT 20
	`

	rows, err := h.db.Queryx(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	for rows.Next() {
		stat := make(map[string]interface{})
		err := rows.MapScan(stat)
		if err != nil {
			continue
		}
		stats = append(stats, stat)
	}

	c.JSON(http.StatusOK, stats)
}
