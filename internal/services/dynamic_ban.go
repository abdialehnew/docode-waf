package services

import (
	"context"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/redis/go-redis/v9"
)

const (
	defaultBanThreshold = 10
	defaultBanWindow    = 60 * time.Second
	defaultBanDuration  = 1 * time.Hour
	violationKeyPrefix  = "waf:violation:"
	bannedKeyPrefix     = "waf:banned:"
)

// DynamicBanService manages dynamic IP bans
type DynamicBanService struct {
	db                  *sqlx.DB
	redisClient         *redis.Client
	notificationService *NotificationService
	threshold           int
	window              time.Duration
	banDuration         time.Duration
}

// NewDynamicBanService creates a new dynamic ban service
func NewDynamicBanService(db *sqlx.DB, redisClient *redis.Client, notificationService *NotificationService) *DynamicBanService {
	// TODO: Load config from environment variables
	return &DynamicBanService{
		db:                  db,
		redisClient:         redisClient,
		notificationService: notificationService,
		threshold:           defaultBanThreshold,
		window:              defaultBanWindow,
		banDuration:         defaultBanDuration,
	}
}

// RecordViolation increments the violation count for an IP
// If count exceeds threshold, bans the IP
func (s *DynamicBanService) RecordViolation(ip string, violationType string) error {
	ctx := context.Background()
	key := violationKeyPrefix + ip

	// Increment violation count
	count, err := s.redisClient.Incr(ctx, key).Result()
	if err != nil {
		return fmt.Errorf("failed to increment violation count: %w", err)
	}

	// Set expiration on first violation
	if count == 1 {
		s.redisClient.Expire(ctx, key, s.window)
	}

	// Check threshold
	if int(count) >= s.threshold {
		reason := fmt.Sprintf("Too many violations (%d) in %s. Last violation: %s", count, s.window, violationType)
		return s.BanIP(ip, reason, s.banDuration)
	}

	return nil
}

// BanIP bans an IP address
func (s *DynamicBanService) BanIP(ip string, reason string, duration time.Duration) error {
	ctx := context.Background()

	// 1. Add to Redis for fast lookups
	bannedKey := bannedKeyPrefix + ip
	err := s.redisClient.Set(ctx, bannedKey, reason, duration).Err()
	if err != nil {
		return fmt.Errorf("failed to cache ban in redis: %w", err)
	}

	// 2. Persist to Database
	query := `
		INSERT INTO ip_bans (ip_address, reason, expires_at)
		VALUES ($1, $2, $3)
	`
	expiresAt := time.Now().Add(duration)
	_, err = s.db.Exec(query, ip, reason, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to persist ban to db: %w", err)
	}

	// 3. Clear violation counter
	s.redisClient.Del(ctx, violationKeyPrefix+ip)

	fmt.Printf("[DynamicBan] Banned IP %s for %s. Reason: %s\n", ip, duration, reason)

	// Send notification
	if s.notificationService != nil {
		s.notificationService.Notify(NotificationEvent{
			Type:      "ip_banned",
			Title:     "IP Address Banned",
			Message:   fmt.Sprintf("IP %s has been banned for %s.\nReason: %s", ip, duration, reason),
			Severity:  "medium",
			Timestamp: time.Now(),
			Metadata: map[string]interface{}{
				"IP":       ip,
				"Reason":   reason,
				"Duration": duration.String(),
			},
		})
	}

	return nil
}

// IsBanned checks if an IP is currently banned
func (s *DynamicBanService) IsBanned(ip string) (bool, string) {
	ctx := context.Background()
	key := bannedKeyPrefix + ip

	// Check Redis first
	reason, err := s.redisClient.Get(ctx, key).Result()
	if err == nil {
		return true, reason
	}
	if err != redis.Nil {
		// Log error but continue to DB check? Or fail safe?
		// For now fail open check DB
	}

	// Check DB (fallback)
	// Theoretically we should sync DB bans to Redis on startup or cache miss
	// But for "Dynamic Bans" usually Redis is the source of truth for short term
	// Let's rely on Redis for high performance, DB for history/audit

	return false, ""
}

// UnbanIP removes a ban
func (s *DynamicBanService) UnbanIP(ip string) error {
	ctx := context.Background()

	// Remove from Redis
	s.redisClient.Del(ctx, bannedKeyPrefix+ip)
	s.redisClient.Del(ctx, violationKeyPrefix+ip)

	// Update DB (mark as expired or delete?)
	// Let's just update expires_at to now
	query := `UPDATE ip_bans SET expires_at = NOW() WHERE ip_address = $1 AND expires_at > NOW()`
	_, err := s.db.Exec(query, ip)
	return err
}
