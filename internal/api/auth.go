package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"

	"github.com/aleh/docode-waf/internal/config"
	"github.com/aleh/docode-waf/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	authService  *services.AuthService
	emailService *services.EmailService
	cfg          *config.Config
	db           *sqlx.DB
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *services.AuthService, emailService *services.EmailService, cfg *config.Config, db *sqlx.DB) *AuthHandler {
	return &AuthHandler{
		authService:  authService,
		emailService: emailService,
		cfg:          cfg,
		db:           db,
	}
}

// isTurnstileEnabledForLogin checks if Turnstile is enabled for login from database
func (h *AuthHandler) isTurnstileEnabledForLogin() bool {
	var settings struct {
		TurnstileEnabled      bool `db:"turnstile_enabled"`
		TurnstileLoginEnabled bool `db:"turnstile_login_enabled"`
	}
	query := `
		SELECT 
			COALESCE(turnstile_enabled, false) as turnstile_enabled,
			COALESCE(turnstile_login_enabled, false) as turnstile_login_enabled
		FROM app_settings 
		WHERE id = 1
	`
	err := h.db.Get(&settings, query)
	if err != nil {
		return false // Default to disabled if error
	}
	return settings.TurnstileEnabled && settings.TurnstileLoginEnabled
}

// isTurnstileEnabledForRegister checks if Turnstile is enabled for register from database
func (h *AuthHandler) isTurnstileEnabledForRegister() bool {
	var settings struct {
		TurnstileEnabled         bool `db:"turnstile_enabled"`
		TurnstileRegisterEnabled bool `db:"turnstile_register_enabled"`
	}
	query := `
		SELECT 
			COALESCE(turnstile_enabled, false) as turnstile_enabled,
			COALESCE(turnstile_register_enabled, false) as turnstile_register_enabled
		FROM app_settings 
		WHERE id = 1
	`
	err := h.db.Get(&settings, query)
	if err != nil {
		return false // Default to disabled if error
	}
	return settings.TurnstileEnabled && settings.TurnstileRegisterEnabled
}

// Register creates a new super admin account
func (h *AuthHandler) Register(c *gin.Context) {
	var input struct {
		Username       string `json:"username" binding:"required,min=3,max=50"`
		Email          string `json:"email" binding:"required,email"`
		Password       string `json:"password" binding:"required,min=8"`
		FullName       string `json:"full_name" binding:"required"`
		TurnstileToken string `json:"turnstile_token"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify Turnstile token if configured AND enabled in database
	if h.cfg.Turnstile.SecretKey != "" && h.cfg.Turnstile.SecretKey != "${TURNSTILE_SECRET_KEY}" && h.isTurnstileEnabledForRegister() {
		if !h.verifyTurnstile(input.TurnstileToken, c.ClientIP()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid captcha verification"})
			return
		}
	}

	admin, err := h.authService.Register(input.Username, input.Email, input.Password, input.FullName)
	if err != nil {
		if err == services.ErrUserExists {
			c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create admin"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Admin registered successfully",
		"admin":   admin,
	})
}

// Login authenticates an admin and returns a JWT token
func (h *AuthHandler) Login(c *gin.Context) {
	var input struct {
		Username       string `json:"username" binding:"required"`
		Password       string `json:"password" binding:"required"`
		TurnstileToken string `json:"turnstile_token"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify Turnstile token if configured AND enabled in database
	if h.cfg.Turnstile.SecretKey != "" && h.cfg.Turnstile.SecretKey != "${TURNSTILE_SECRET_KEY}" && h.isTurnstileEnabledForLogin() {
		if !h.verifyTurnstile(input.TurnstileToken, c.ClientIP()) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid captcha verification"})
			return
		}
	}

	token, admin, err := h.authService.Login(input.Username, input.Password)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid username or password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Login failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"token":   token,
		"admin":   admin,
	})
}

// RequestPasswordReset initiates password reset process
func (h *AuthHandler) RequestPasswordReset(c *gin.Context) {
	var input struct {
		Email string `json:"email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	username, token, err := h.authService.RequestPasswordReset(input.Email)
	if err != nil {
		// Don't reveal if email exists for security
		c.JSON(http.StatusOK, gin.H{
			"message": "If the email exists, a reset link will be sent",
		})
		return
	}

	// Send password reset email
	err = h.emailService.SendPasswordResetEmail(input.Email, username, token)
	if err != nil {
		// Log error but don't reveal to user
		// In production, log this: log.Printf("Failed to send reset email: %v", err)
		c.JSON(http.StatusOK, gin.H{
			"message": "If the email exists, a reset link will be sent",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a password reset email has been sent",
	})
}

// ResetPassword resets password using reset token
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var input struct {
		Token       string `json:"token" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := h.authService.ResetPassword(input.Token, input.NewPassword)
	if err != nil {
		if err == services.ErrInvalidToken {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reset password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successful",
	})
}

// ChangePassword changes user password (authenticated route)
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var input struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get admin ID from context (set by auth middleware)
	adminID, exists := c.Get("admin_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	err := h.authService.ChangePassword(adminID.(string), input.OldPassword, input.NewPassword)
	if err != nil {
		if err == services.ErrInvalidCredentials {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid old password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to change password"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password changed successfully",
	})
}

// GetProfile returns current admin profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	admin, exists := c.Get("admin")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	c.JSON(http.StatusOK, admin)
}

// verifyTurnstile verifies Cloudflare Turnstile token
func (h *AuthHandler) verifyTurnstile(token, remoteIP string) bool {
	if token == "" {
		return false
	}

	// Prepare verification request
	formData := url.Values{}
	formData.Set("secret", h.cfg.Turnstile.SecretKey)
	formData.Set("response", token)
	formData.Set("remoteip", remoteIP)

	// Send verification request to Cloudflare
	resp, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify", formData)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false
	}

	// Parse response
	var result struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return false
	}

	return result.Success
}
