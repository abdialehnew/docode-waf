package api

import (
	"net/http"
	"time"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)
type DashboardHandler struct {
	db *sqlx.DB
}
func NewDashboardHandler(db *sqlx.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
type DashboardStats struct {
	TotalRequests   int64              `json:"total_requests"`
	BlockedRequests int64              `json:"blocked_requests"`
	TotalAttacks    int64              `json:"total_attacks"`
	UniqueIPs       int64              `json:"unique_ips"`
	AvgResponseTime float64            `json:"avg_response_time"`
	RequestsByHour  []RequestByHour    `json:"requests_by_hour"`
	TopCountries    []CountryStats     `json:"top_countries"`
	TopAttackTypes  []AttackTypeStats  `json:"top_attack_types"`
	RecentAttacks   []RecentAttack     `json:"recent_attacks"`
type RequestByHour struct {
	Hour     time.Time `json:"hour" db:"hour"`
	Count    int64     `json:"count" db:"count"`
	Blocked  int64     `json:"blocked" db:"blocked"`
type CountryStats struct {
	Country string `json:"country" db:"country_code"`
	Count   int64  `json:"count" db:"count"`
type AttackTypeStats struct {
	Type  string `json:"type" db:"attack_type"`
	Count int64  `json:"count" db:"count"`
}	c.JSON(http.StatusOK, logs)	}		return		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})	if err != nil {	`, limit, offset)		LIMIT $1 OFFSET $2		ORDER BY timestamp DESC		FROM traffic_logs		SELECT timestamp, client_ip, method, url, status_code, response_time, blocked	err := h.db.Select(&logs, `	}		Blocked      bool      `json:"blocked" db:"blocked"`		ResponseTime int       `json:"response_time" db:"response_time"`		StatusCode   int       `json:"status_code" db:"status_code"`		URL          string    `json:"url" db:"url"`		Method       string    `json:"method" db:"method"`		ClientIP     string    `json:"client_ip" db:"client_ip"`		Timestamp    time.Time `json:"timestamp" db:"timestamp"`	var logs []struct {	offset := c.DefaultQuery("offset", "0")	limit := c.DefaultQuery("limit", "100")func (h *DashboardHandler) GetTrafficLogs(c *gin.Context) {}	c.JSON(http.StatusOK, stats)	`, since)		LIMIT 20		ORDER BY timestamp DESC		WHERE timestamp > $1		FROM attack_logs		SELECT timestamp, client_ip, attack_type, severity, description	h.db.Select(&stats.RecentAttacks, `	// Recent attacks	`, since)		LIMIT 10		ORDER BY count DESC		GROUP BY attack_type		WHERE timestamp > $1		FROM attack_logs		SELECT attack_type, COUNT(*) as count	h.db.Select(&stats.TopAttackTypes, `	// Top attack types	`, since)		LIMIT 10		ORDER BY count DESC		GROUP BY country_code		WHERE timestamp > $1 AND country_code != ''		FROM traffic_logs		SELECT country_code, COUNT(*) as count	h.db.Select(&stats.TopCountries, `	// Top countries	`, since)		LIMIT 24		ORDER BY hour DESC		GROUP BY hour		WHERE timestamp > $1		FROM traffic_logs			SUM(CASE WHEN blocked THEN 1 ELSE 0 END) as blocked			COUNT(*) as count,			date_trunc('hour', timestamp) as hour,		SELECT 	h.db.Select(&stats.RequestsByHour, `	// Requests by hour		"SELECT COALESCE(AVG(response_time), 0) FROM traffic_logs WHERE timestamp > $1", since)	h.db.Get(&stats.AvgResponseTime,	// Average response time		"SELECT COUNT(DISTINCT client_ip) FROM traffic_logs WHERE timestamp > $1", since)	h.db.Get(&stats.UniqueIPs,	// Unique IPs		"SELECT COUNT(*) FROM attack_logs WHERE timestamp > $1", since)	h.db.Get(&stats.TotalAttacks,	// Total attacks		"SELECT COUNT(*) FROM traffic_logs WHERE timestamp > $1 AND blocked = true", since)	h.db.Get(&stats.BlockedRequests,	// Blocked requests		"SELECT COUNT(*) FROM traffic_logs WHERE timestamp > $1", since)	h.db.Get(&stats.TotalRequests, 	// Total requests	stats := DashboardStats{}	since := time.Now().Add(-duration)	}		duration = 30 * 24 * time.Hour	case "30d":		duration = 7 * 24 * time.Hour	case "7d":		duration = 24 * time.Hour	case "24h":		duration = 6 * time.Hour	case "6h":		duration = 1 * time.Hour	case "1h":	switch timeRange {	duration := 24 * time.Hour		timeRange := c.DefaultQuery("range", "24h")func (h *DashboardHandler) GetStats(c *gin.Context) {}	Description string    `json:"description" db:"description"`	Severity    string    `json:"severity" db:"severity"`	AttackType  string    `json:"attack_type" db:"attack_type"`	ClientIP    string    `json:"client_ip" db:"client_ip"`	Timestamp   time.Time `json:"timestamp" db:"timestamp"`type RecentAttack struct {
