package api

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

const (
	nginxLogsDir      = "/app/nginx/logs"
	errDomainRequired = "domain parameter is required"
	defaultLogLines   = "100"
	defaultLiveMode   = "false"
)

// LogsHandler handles log viewing requests
type LogsHandler struct {
	db *sqlx.DB
}

// NewLogsHandler creates a new logs handler
func NewLogsHandler(db *sqlx.DB) *LogsHandler {
	return &LogsHandler{db: db}
}

// GetNginxAccessLogs returns nginx access logs for a specific vhost
func (h *LogsHandler) GetNginxAccessLogs(c *gin.Context) {
	domain := c.Query("domain")
	lines := c.DefaultQuery("lines", defaultLogLines)
	live := c.DefaultQuery("live", defaultLiveMode)

	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errDomainRequired,
		})
		return
	}

	logPath := filepath.Join(nginxLogsDir, domain, "access.log")
	logs, err := readLogFile(logPath, lines)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read access log",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domain": domain,
		"type":   "access",
		"lines":  logs,
		"live":   live == "true",
	})
}

// GetNginxErrorLogs returns nginx error logs for a specific vhost
func (h *LogsHandler) GetNginxErrorLogs(c *gin.Context) {
	domain := c.Query("domain")
	lines := c.DefaultQuery("lines", defaultLogLines)
	live := c.DefaultQuery("live", defaultLiveMode)

	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errDomainRequired,
		})
		return
	}

	logPath := filepath.Join(nginxLogsDir, domain, "error.log")
	logs, err := readLogFile(logPath, lines)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to read error log",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"domain": domain,
		"type":   "error",
		"lines":  logs,
		"live":   live == "true",
	})
}

// GetWAFLogs returns WAF logs from database
func (h *LogsHandler) GetWAFLogs(c *gin.Context) {
	startDate := c.DefaultQuery("start_date", time.Now().Add(-24*time.Hour).Format("2006-01-02"))
	endDate := c.DefaultQuery("end_date", time.Now().Format("2006-01-02"))
	limit := c.DefaultQuery("limit", "100")
	isAttack := c.Query("is_attack")

	query := `
		SELECT id, timestamp, client_ip, method, url, status_code, 
		       response_time, user_agent, blocked, block_reason, 
		       country_code, is_attack, attack_type, host
		FROM traffic_logs
		WHERE timestamp >= $1::date AND timestamp < ($2::date + interval '1 day')
	`

	args := []interface{}{startDate, endDate}

	if isAttack != "" {
		query += ` AND is_attack = $3`
		args = append(args, isAttack == "true")
		query += ` ORDER BY timestamp DESC LIMIT $4`
	} else {
		query += ` ORDER BY timestamp DESC LIMIT $3`
	}

	args = append(args, limit)

	var logs []struct {
		ID           string    `db:"id" json:"id"`
		Timestamp    time.Time `db:"timestamp" json:"timestamp"`
		ClientIP     string    `db:"client_ip" json:"client_ip"`
		Method       string    `db:"method" json:"method"`
		URL          string    `db:"url" json:"url"`
		StatusCode   int       `db:"status_code" json:"status_code"`
		ResponseTime int       `db:"response_time" json:"response_time"`
		UserAgent    string    `db:"user_agent" json:"user_agent"`
		Blocked      bool      `db:"blocked" json:"blocked"`
		BlockReason  *string   `db:"block_reason" json:"block_reason"`
		CountryCode  *string   `db:"country_code" json:"country_code"`
		IsAttack     bool      `db:"is_attack" json:"is_attack"`
		AttackType   *string   `db:"attack_type" json:"attack_type"`
		Host         string    `db:"host" json:"host"`
	}

	if err := h.db.Select(&logs, query, args...); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch WAF logs",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"logs":       logs,
		"start_date": startDate,
		"end_date":   endDate,
		"total":      len(logs),
	})
}

// StreamNginxLogs streams nginx logs in real-time (SSE)
func (h *LogsHandler) StreamNginxLogs(c *gin.Context) {
	domain := c.Query("domain")
	logType := c.DefaultQuery("type", "access")

	if domain == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errDomainRequired,
		})
		return
	}

	filename := logType + ".log"
	logPath := filepath.Join(nginxLogsDir, domain, filename)

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Stream last 10 lines immediately
	logs, _ := readLogFile(logPath, "10")
	for _, line := range logs {
		c.SSEvent("message", line)
		c.Writer.Flush()
	}

	// Poll for new lines every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case <-ticker.C:
			// Check for new lines
			newLogs, err := readLogFile(logPath, "1")
			if err == nil && len(newLogs) > 0 {
				c.SSEvent("message", newLogs[0])
				c.Writer.Flush()
			}
		}
	}
}

// Helper function to read log file
func readLogFile(path string, lines string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var logLines []string
	scanner := bufio.NewScanner(file)

	// Read all lines first
	for scanner.Scan() {
		logLines = append(logLines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Return last N lines
	n := 100
	if lines != "" {
		if parsed := parseInt(lines); parsed > 0 {
			n = parsed
		}
	}

	if len(logLines) > n {
		return logLines[len(logLines)-n:], nil
	}

	return logLines, nil
}

func parseInt(s string) int {
	var n int
	for _, ch := range s {
		if ch >= '0' && ch <= '9' {
			n = n*10 + int(ch-'0')
		} else {
			return 0
		}
	}
	return n
}

// GetVHostsForLogs returns list of vhosts for log filtering
func (h *LogsHandler) GetVHostsForLogs(c *gin.Context) {
	query := `
		SELECT domain, name 
		FROM vhosts 
		WHERE enabled = true 
		ORDER BY domain
	`

	var vhosts []struct {
		Domain string `db:"domain" json:"domain"`
		Name   string `db:"name" json:"name"`
	}

	if err := h.db.Select(&vhosts, query); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch vhosts",
		})
		return
	}

	// Return all enabled vhosts with log file status
	var vhostsList []map[string]interface{}
	for _, vhost := range vhosts {
		accessLogPath := filepath.Join(nginxLogsDir, vhost.Domain, "access.log")
		errorLogPath := filepath.Join(nginxLogsDir, vhost.Domain, "error.log")

		hasAccessLog := fileExists(accessLogPath)
		hasErrorLog := fileExists(errorLogPath)

		vhostsList = append(vhostsList, map[string]interface{}{
			"domain":         vhost.Domain,
			"name":           vhost.Name,
			"has_access_log": hasAccessLog,
			"has_error_log":  hasErrorLog,
		})
	}

	c.JSON(http.StatusOK, vhostsList)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
