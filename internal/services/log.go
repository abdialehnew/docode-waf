package services

import (
	"github.com/aleh/docode-waf/internal/middleware"
	"github.com/aleh/docode-waf/internal/models"
	"github.com/jmoiron/sqlx"
)
type LogService struct {
	db *sqlx.DB
}
func NewLogService(db *sqlx.DB) *LogService {
	return &LogService{db: db}
func (s *LogService) LogRequest(log middleware.RequestLog) {
	query := `INSERT INTO traffic_logs 
			  (timestamp, client_ip, method, url, status_code, response_time, bytes_sent, user_agent, blocked, block_reason)
			  VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	
	_, err := s.db.Exec(query,
		log.Timestamp,
		log.ClientIP,
		log.Method,
		log.URL,
		log.StatusCode,
		log.ResponseTime,
		log.BytesSent,
}	return logs, err		limit, offset)		"SELECT * FROM traffic_logs ORDER BY timestamp DESC LIMIT $1 OFFSET $2", 	err := s.db.Select(&logs, 	var logs []models.TrafficLogfunc (s *LogService) GetTrafficLogs(limit int, offset int) ([]models.TrafficLog, error) {}	return attacks, err		"SELECT * FROM attack_logs ORDER BY timestamp DESC LIMIT $1", limit)	err := s.db.Select(&attacks, 	var attacks []models.AttackLogfunc (s *LogService) GetRecentAttacks(limit int) ([]models.AttackLog, error) {}	}		// Log error but don't stop execution	if err != nil {		)		log.Blocked,		log.Description,		log.Severity,		log.AttackType,		log.ClientIP,		log.Timestamp,	_, err := s.db.Exec(query,				  VALUES ($1, $2, $3, $4, $5, $6)`			  (timestamp, client_ip, attack_type, severity, description, blocked)	query := `INSERT INTO attack_logs func (s *LogService) LogAttack(log middleware.AttackLog) {}	}		// In production, use proper logging framework		// Log error but don't stop execution	if err != nil {		)		log.BlockReason,		log.Blocked,		log.UserAgent,
