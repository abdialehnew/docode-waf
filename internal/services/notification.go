package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jmoiron/sqlx"
)

// NotificationChannel represents a destination for alerts
type NotificationChannel struct {
	ID        string          `db:"id" json:"id"`
	Name      string          `db:"name" json:"name"`
	Type      string          `db:"type" json:"type"`     // email, slack, discord, webhook
	Config    json.RawMessage `db:"config" json:"config"` // {"webhook_url": "...", "email": "..."}
	Events    json.RawMessage `db:"events" json:"events"` // ["attack_detected", "ip_banned"]
	Enabled   bool            `db:"enabled" json:"enabled"`
	CreatedAt time.Time       `db:"created_at" json:"created_at"`
	UpdatedAt time.Time       `db:"updated_at" json:"updated_at"`
}

type NotificationService struct {
	db           *sqlx.DB
	emailService *EmailService
}

func NewNotificationService(db *sqlx.DB, emailService *EmailService) *NotificationService {
	return &NotificationService{
		db:           db,
		emailService: emailService,
	}
}

// NotificationEvent represents the data payload for an alert
type NotificationEvent struct {
	Type      string                 `json:"type"`     // attack_detected, ip_banned, admin_login
	Title     string                 `json:"title"`    // e.g. "Critical SQL Injection Detected"
	Message   string                 `json:"message"`  // Human readable body
	Severity  string                 `json:"severity"` // critical, high, medium, low, info
	Metadata  map[string]interface{} `json:"metadata"` // Extra fields (IP, URL, UserAgent)
	Timestamp time.Time              `json:"timestamp"`
}

// Notify dispatches an event to all subscribed and enabled channels
func (s *NotificationService) Notify(event NotificationEvent) {
	// Set timestamp if missing
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Fetch all enabled channels
	var channels []NotificationChannel
	err := s.db.Select(&channels, "SELECT * FROM notification_channels WHERE enabled = true")
	if err != nil {
		log.Printf("Failed to fetch notification channels: %v", err)
		return
	}

	for _, channel := range channels {
		if s.isSubscribed(channel, event.Type) {
			go s.sendToChannelInternal(channel, event)
		}
	}
}

// isSubscribed checks if the channel subscribes to the event type
func (s *NotificationService) isSubscribed(channel NotificationChannel, eventType string) bool {
	var subscribedEvents []string
	if err := json.Unmarshal(channel.Events, &subscribedEvents); err != nil {
		return false
	}

	for _, e := range subscribedEvents {
		if e == eventType || e == "*" {
			return true
		}
	}
	return false
}

// SendToChannel routes the event to the specific provider using the provided config
func (s *NotificationService) SendToChannel(channel *NotificationChannel, event NotificationEvent) error {
	var config map[string]string
	if err := json.Unmarshal(channel.Config, &config); err != nil {
		return fmt.Errorf("invalid config: %v", err)
	}

	switch channel.Type {
	case "slack":
		return s.sendToSlack(config["webhook_url"], event)
	case "discord":
		return s.sendToDiscord(config["webhook_url"], event)
	case "webhook":
		return s.sendToWebhook(config["webhook_url"], event)
	case "email":
		return s.sendToEmail(config["email_address"], event)
	}
	return fmt.Errorf("unknown channel type: %s", channel.Type)
}

// sendToChannelInternal is a wrapper for async processing that logs errors
func (s *NotificationService) sendToChannelInternal(channel NotificationChannel, event NotificationEvent) {
	if err := s.SendToChannel(&channel, event); err != nil {
		log.Printf("Failed to send notification to %s (%s): %v", channel.Name, channel.Type, err)
	}
}

// === Providers ===

func (s *NotificationService) sendToSlack(url string, event NotificationEvent) error {
	color := "#36a64f" // Green
	if event.Severity == "critical" || event.Severity == "high" {
		color = "#ff0000" // Red
	} else if event.Severity == "medium" {
		color = "#ffa500" // Orange
	}

	payload := map[string]interface{}{
		"attachments": []map[string]interface{}{
			{
				"color": color,
				"title": event.Title,
				"text":  event.Message,
				"fields": []map[string]string{
					{"title": "Event Type", "value": event.Type, "short": "true"},
					{"title": "Severity", "value": event.Severity, "short": "true"},
				},
				"ts": event.Timestamp.Unix(),
			},
		},
	}

	// Add metadata to fields
	for k, v := range event.Metadata {
		payload["attachments"].([]map[string]interface{})[0]["fields"] = append(
			payload["attachments"].([]map[string]interface{})[0]["fields"].([]map[string]string),
			map[string]string{"title": k, "value": fmt.Sprintf("%v", v), "short": "true"},
		)
	}

	return postJSON(url, payload)
}

func (s *NotificationService) sendToDiscord(url string, event NotificationEvent) error {
	color := 3066993 // Green
	if event.Severity == "critical" || event.Severity == "high" {
		color = 15158332 // Red
	} else if event.Severity == "medium" {
		color = 16753920 // Orange
	}

	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{
			{
				"title":       event.Title,
				"description": event.Message,
				"color":       color,
				"fields": []map[string]string{
					{"name": "Event Type", "value": event.Type, "inline": "true"},
					{"name": "Severity", "value": event.Severity, "inline": "true"},
				},
				"timestamp": event.Timestamp.Format(time.RFC3339),
			},
		},
	}

	// Add metadata
	for k, v := range event.Metadata {
		payload["embeds"].([]map[string]interface{})[0]["fields"] = append(
			payload["embeds"].([]map[string]interface{})[0]["fields"].([]map[string]string),
			map[string]string{"name": k, "value": fmt.Sprintf("%v", v), "inline": "true"},
		)
	}

	return postJSON(url, payload)
}

func (s *NotificationService) sendToWebhook(url string, event NotificationEvent) error {
	return postJSON(url, event)
}

func (s *NotificationService) sendToEmail(email string, event NotificationEvent) error {
	if s.emailService == nil {
		return fmt.Errorf("email service not initialized")
	}

	subject := fmt.Sprintf("[%s] %s", event.Severity, event.Title)

	// Simple HTML body
	body := fmt.Sprintf("<h2>%s</h2><p>%s</p>", event.Title, event.Message)
	body += "<h3>Details:</h3><ul>"
	for k, v := range event.Metadata {
		body += fmt.Sprintf("<li><strong>%s:</strong> %v</li>", k, v)
	}
	body += "</ul>"
	body += fmt.Sprintf("<p><small>Time: %s</small></p>", event.Timestamp.Format(time.RFC1123))

	return s.emailService.SendEmail(email, subject, body)
}

// Helper to post JSON
func postJSON(url string, payload interface{}) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(data))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("server returned status: %d", resp.StatusCode)
	}

	return nil
}
