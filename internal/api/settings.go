package api

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

type SettingsHandler struct {
	db *sqlx.DB
}

func NewSettingsHandler(db *sqlx.DB) *SettingsHandler {
	return &SettingsHandler{db: db}
}

type AppSettings struct {
	AppName       string  `json:"app_name" db:"app_name"`
	AppLogo       string  `json:"app_logo" db:"app_logo"`
	SignupEnabled bool    `json:"signup_enabled" db:"signup_enabled"`
	SMTPHost      *string `json:"smtp_host" db:"smtp_host"`
	SMTPPort      *int    `json:"smtp_port" db:"smtp_port"`
	SMTPUsername  *string `json:"smtp_username" db:"smtp_username"`
	SMTPPassword  *string `json:"smtp_password" db:"smtp_password"`
	SMTPFromEmail *string `json:"smtp_from_email" db:"smtp_from_email"`
	SMTPFromName  *string `json:"smtp_from_name" db:"smtp_from_name"`
	SMTPUseTLS    bool    `json:"smtp_use_tls" db:"smtp_use_tls"`
}

// GetAppSettings retrieves application settings
func (h *SettingsHandler) GetAppSettings(c *gin.Context) {
	var settings AppSettings

	query := `
		SELECT 
			app_name, app_logo, signup_enabled,
			smtp_host, smtp_port, smtp_username, smtp_password,
			smtp_from_email, smtp_from_name, smtp_use_tls
		FROM app_settings 
		WHERE id = 1
	`
	err := h.db.Get(&settings, query)

	if err != nil {
		if err == sql.ErrNoRows {
			// Return default values if no settings exist
			defaultPort := 587
			defaultFromName := "Docode WAF"
			c.JSON(http.StatusOK, AppSettings{
				AppName:       "Docode WAF",
				AppLogo:       "",
				SignupEnabled: true,
				SMTPPort:      &defaultPort,
				SMTPFromName:  &defaultFromName,
				SMTPUseTLS:    true,
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, settings)
}

// SaveAppSettings saves or updates application settings
func (h *SettingsHandler) SaveAppSettings(c *gin.Context) {
	var settings AppSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate
	if settings.AppName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Application name is required"})
		return
	}

	// Check if settings exist
	var exists bool
	err := h.db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM app_settings WHERE id = 1)`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if exists {
		// Update existing settings
		query := `
			UPDATE app_settings 
			SET app_name = $1, app_logo = $2, signup_enabled = $3,
				smtp_host = $4, smtp_port = $5, smtp_username = $6, smtp_password = $7,
				smtp_from_email = $8, smtp_from_name = $9, smtp_use_tls = $10,
				updated_at = NOW() 
			WHERE id = 1
		`
		_, err = h.db.Exec(query,
			settings.AppName, settings.AppLogo, settings.SignupEnabled,
			settings.SMTPHost, settings.SMTPPort, settings.SMTPUsername, settings.SMTPPassword,
			settings.SMTPFromEmail, settings.SMTPFromName, settings.SMTPUseTLS)
	} else {
		// Insert new settings
		query := `
			INSERT INTO app_settings (
				id, app_name, app_logo, signup_enabled,
				smtp_host, smtp_port, smtp_username, smtp_password,
				smtp_from_email, smtp_from_name, smtp_use_tls,
				created_at, updated_at
			) VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW(), NOW())
		`
		_, err = h.db.Exec(query,
			settings.AppName, settings.AppLogo, settings.SignupEnabled,
			settings.SMTPHost, settings.SMTPPort, settings.SMTPUsername, settings.SMTPPassword,
			settings.SMTPFromEmail, settings.SMTPFromName, settings.SMTPUseTLS)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings saved successfully"})
}
