package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/aleh/docode-waf/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type NotificationsHandler struct {
	db                  *sqlx.DB
	notificationService *services.NotificationService
}

func NewNotificationsHandler(db *sqlx.DB, notificationService *services.NotificationService) *NotificationsHandler {
	return &NotificationsHandler{
		db:                  db,
		notificationService: notificationService,
	}
}

// ListChannels returns all notification channels
func (h *NotificationsHandler) ListChannels(c *gin.Context) {
	var channels []services.NotificationChannel
	err := h.db.Select(&channels, "SELECT * FROM notification_channels ORDER BY created_at DESC")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notification channels"})
		return
	}

	if channels == nil {
		channels = []services.NotificationChannel{}
	}

	c.JSON(http.StatusOK, channels)
}

// GetChannel returns a specific channel
func (h *NotificationsHandler) GetChannel(c *gin.Context) {
	id := c.Param("id")
	var channel services.NotificationChannel
	err := h.db.Get(&channel, "SELECT * FROM notification_channels WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}
	c.JSON(http.StatusOK, channel)
}

// CreateChannel creates a new notification channel
func (h *NotificationsHandler) CreateChannel(c *gin.Context) {
	var req struct {
		Name    string          `json:"name" binding:"required"`
		Type    string          `json:"type" binding:"required"`
		Config  json.RawMessage `json:"config" binding:"required"`
		Events  json.RawMessage `json:"events" binding:"required"`
		Enabled bool            `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var id string
	query := `
		INSERT INTO notification_channels (name, type, config, events, enabled)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	// Convert json.RawMessage to string for JSONB compatibility
	err := h.db.QueryRow(query, req.Name, req.Type, string(req.Config), string(req.Events), req.Enabled).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id, "message": "Channel created successfully"})
}

// UpdateChannel updates an existing channel
func (h *NotificationsHandler) UpdateChannel(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name    string          `json:"name"`
		Type    string          `json:"type"`
		Config  json.RawMessage `json:"config"`
		Events  json.RawMessage `json:"events"`
		Enabled *bool           `json:"enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update only provided fields
	// Handle complex COALESCE logic for JSONB in application code simply or assume full update for simplicity if frontend sends everything
	// Let's rely on standard full update for this example since frontend usually sends full object

	fullQuery := `
		UPDATE notification_channels
		SET name = $2, type = $3, config = $4, events = $5, enabled = $6
		WHERE id = $1
	`
	// For partial update support (PATCH-like behavior via PUT if frontend passes current state)
	// Here assume frontend sends complete data for PUT

	_, err := h.db.Exec(fullQuery, id, req.Name, req.Type, string(req.Config), string(req.Events), req.Enabled)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update channel: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Channel updated successfully"})
}

// DeleteChannel removes a channel
func (h *NotificationsHandler) DeleteChannel(c *gin.Context) {
	id := c.Param("id")
	_, err := h.db.Exec("DELETE FROM notification_channels WHERE id = $1", id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete channel"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Channel deleted successfully"})
}

// TestNotification sends a test alert to a specific channel configuration
// It does NOT require saving the channel first (allows verifying config)
func (h *NotificationsHandler) TestNotification(c *gin.Context) {
	var req struct {
		Type   string          `json:"type" binding:"required"`
		Config json.RawMessage `json:"config" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Create a temporary channel object
	channel := services.NotificationChannel{
		Name:   "Test Channel",
		Type:   req.Type,
		Config: req.Config,
	}

	testEvent := services.NotificationEvent{
		Type:      "test_event",
		Title:     "Test Notification",
		Message:   "This is a test alert to verify your notification settings. If you see this, it works!",
		Severity:  "info",
		Timestamp: time.Now(),
		Metadata:  map[string]interface{}{"User": "Admin", "Action": "Test"},
	}

	if err := h.notificationService.SendToChannel(&channel, testEvent); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to send test notification: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Test notification sent successfully"})
}
