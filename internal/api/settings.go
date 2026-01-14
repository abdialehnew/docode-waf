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
	AppName                  string  `json:"app_name" db:"app_name"`
	AppLogo                  string  `json:"app_logo" db:"app_logo"`
	SignupEnabled            bool    `json:"signup_enabled" db:"signup_enabled"`
	SMTPHost                 *string `json:"smtp_host" db:"smtp_host"`
	SMTPPort                 *int    `json:"smtp_port" db:"smtp_port"`
	SMTPUsername             *string `json:"smtp_username" db:"smtp_username"`
	SMTPPassword             *string `json:"smtp_password" db:"smtp_password"`
	SMTPFromEmail            *string `json:"smtp_from_email" db:"smtp_from_email"`
	SMTPFromName             *string `json:"smtp_from_name" db:"smtp_from_name"`
	SMTPUseTLS               bool    `json:"smtp_use_tls" db:"smtp_use_tls"`
	TurnstileEnabled         bool    `json:"turnstile_enabled" db:"turnstile_enabled"`
	TurnstileLoginEnabled    bool    `json:"turnstile_login_enabled" db:"turnstile_login_enabled"`
	TurnstileRegisterEnabled bool    `json:"turnstile_register_enabled" db:"turnstile_register_enabled"`
}

// GetAppSettings retrieves application settings
func (h *SettingsHandler) GetAppSettings(c *gin.Context) {
	var settings AppSettings

	query := `
		SELECT 
			app_name, app_logo, signup_enabled,
			smtp_host, smtp_port, smtp_username, smtp_password,
			smtp_from_email, smtp_from_name, smtp_use_tls,
			COALESCE(turnstile_enabled, true) as turnstile_enabled,
			COALESCE(turnstile_login_enabled, true) as turnstile_login_enabled,
			COALESCE(turnstile_register_enabled, true) as turnstile_register_enabled
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
				AppName:                  "Docode WAF",
				AppLogo:                  "",
				SignupEnabled:            true,
				SMTPPort:                 &defaultPort,
				SMTPFromName:             &defaultFromName,
				SMTPUseTLS:               true,
				TurnstileEnabled:         false,
				TurnstileLoginEnabled:    false,
				TurnstileRegisterEnabled: false,
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
				turnstile_enabled = $11, turnstile_login_enabled = $12, turnstile_register_enabled = $13,
				updated_at = NOW() 
			WHERE id = 1
		`
		_, err = h.db.Exec(query,
			settings.AppName, settings.AppLogo, settings.SignupEnabled,
			settings.SMTPHost, settings.SMTPPort, settings.SMTPUsername, settings.SMTPPassword,
			settings.SMTPFromEmail, settings.SMTPFromName, settings.SMTPUseTLS,
			settings.TurnstileEnabled, settings.TurnstileLoginEnabled, settings.TurnstileRegisterEnabled)
	} else {
		// Insert new settings
		query := `
			INSERT INTO app_settings (
				id, app_name, app_logo, signup_enabled,
				smtp_host, smtp_port, smtp_username, smtp_password,
				smtp_from_email, smtp_from_name, smtp_use_tls,
				turnstile_enabled, turnstile_login_enabled, turnstile_register_enabled,
				created_at, updated_at
			) VALUES (1, $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, NOW(), NOW())
		`
		_, err = h.db.Exec(query,
			settings.AppName, settings.AppLogo, settings.SignupEnabled,
			settings.SMTPHost, settings.SMTPPort, settings.SMTPUsername, settings.SMTPPassword,
			settings.SMTPFromEmail, settings.SMTPFromName, settings.SMTPUseTLS,
			settings.TurnstileEnabled, settings.TurnstileLoginEnabled, settings.TurnstileRegisterEnabled)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings saved successfully"})
}

// TurnstileSettings contains settings for Turnstile configuration
type TurnstileSettings struct {
	TurnstileEnabled         bool `db:"turnstile_enabled"`
	TurnstileLoginEnabled    bool `db:"turnstile_login_enabled"`
	TurnstileRegisterEnabled bool `db:"turnstile_register_enabled"`
}

// GetTurnstileConfig retrieves Turnstile configuration from database and env
func (h *SettingsHandler) GetTurnstileConfig(c *gin.Context, siteKey string) {
	// Get settings from database
	var settings TurnstileSettings
	query := `
		SELECT 
			COALESCE(turnstile_enabled, false) as turnstile_enabled,
			COALESCE(turnstile_login_enabled, false) as turnstile_login_enabled,
			COALESCE(turnstile_register_enabled, false) as turnstile_register_enabled
		FROM app_settings 
		WHERE id = 1
	`
	err := h.db.Get(&settings, query)

	// Default to disabled if no settings exist or error
	if err != nil {
		settings = TurnstileSettings{
			TurnstileEnabled:         false,
			TurnstileLoginEnabled:    false,
			TurnstileRegisterEnabled: false,
		}
	}

	// Check if site key is configured
	hasSiteKey := siteKey != "" && siteKey != "${TURNSTILE_SITE_KEY}"

	// Turnstile is enabled only if:
	// 1. Site key is configured
	// 2. Global turnstile_enabled is true
	enabled := hasSiteKey && settings.TurnstileEnabled
	loginEnabled := enabled && settings.TurnstileLoginEnabled
	registerEnabled := enabled && settings.TurnstileRegisterEnabled

	c.JSON(http.StatusOK, gin.H{
		"site_key":         siteKey,
		"enabled":          enabled,
		"login_enabled":    loginEnabled,
		"register_enabled": registerEnabled,
	})
}
