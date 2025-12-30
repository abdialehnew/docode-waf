package services

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// LogService handles logging operations
type LogService struct {
	db *sqlx.DB
}

// NewLogService creates a new log service
func NewLogService(db *sqlx.DB) *LogService {
	return &LogService{db: db}
}

// LogAttack logs an attack attempt
func (s *LogService) LogAttack(clientIP, attackType, severity, description string, blocked bool, ruleID *string) error {
	query := `
		INSERT INTO attack_logs (
			id, timestamp, client_ip, attack_type, severity, 
			description, blocked, rule_id
		) VALUES (
			gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7
		)
	`

	_, err := s.db.Exec(query,
		time.Now(),
		clientIP,
		attackType,
		severity,
		description,
		blocked,
		ruleID,
	)

	return err
}

// GetTrafficStats returns traffic statistics
func (s *LogService) GetTrafficStats(startTime, endTime time.Time) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	query := `
		SELECT 
			COUNT(*) as total_requests,
			COUNT(CASE WHEN blocked THEN 1 END) as blocked_requests,
			AVG(response_time) as avg_response_time,
			SUM(bytes_sent) as total_bytes
		FROM traffic_logs
		WHERE timestamp BETWEEN $1 AND $2
	`

	err := s.db.QueryRowx(query, startTime, endTime).MapScan(stats)
	return stats, err
}
