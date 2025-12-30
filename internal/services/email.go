package services

import (
	"crypto/tls"
	"fmt"
	"net/smtp"
	"strings"

	"github.com/jmoiron/sqlx"
)

type SMTPConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
	UseTLS    bool
}

type EmailService struct {
	db *sqlx.DB
}

func NewEmailService(db *sqlx.DB) *EmailService {
	return &EmailService{db: db}
}

// GetSMTPConfig retrieves SMTP configuration from app_settings
func (s *EmailService) GetSMTPConfig() (*SMTPConfig, error) {
	var config SMTPConfig
	query := `
		SELECT 
			COALESCE(smtp_host, '') as host,
			COALESCE(smtp_port, 587) as port,
			COALESCE(smtp_username, '') as username,
			COALESCE(smtp_password, '') as password,
			COALESCE(smtp_from_email, '') as from_email,
			COALESCE(smtp_from_name, 'Docode WAF') as from_name,
			COALESCE(smtp_use_tls, true) as use_tls
		FROM app_settings
		WHERE id = 1
	`
	err := s.db.Get(&config, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get SMTP config: %w", err)
	}

	// Validate required fields
	if config.Host == "" || config.FromEmail == "" {
		return nil, fmt.Errorf("SMTP not configured: host and from_email are required")
	}

	return &config, nil
}

// SendEmail sends an email using configured SMTP settings
func (s *EmailService) SendEmail(to, subject, body string) error {
	config, err := s.GetSMTPConfig()
	if err != nil {
		return err
	}

	// Build message
	from := fmt.Sprintf("%s <%s>", config.FromName, config.FromEmail)
	msg := fmt.Sprintf("From: %s\r\n", from)
	msg += fmt.Sprintf("To: %s\r\n", to)
	msg += fmt.Sprintf("Subject: %s\r\n", subject)
	msg += "MIME-Version: 1.0\r\n"
	msg += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	msg += "\r\n"
	msg += body

	// Connect to SMTP server
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)

	var auth smtp.Auth
	if config.Username != "" && config.Password != "" {
		auth = smtp.PlainAuth("", config.Username, config.Password, config.Host)
	}

	// Send email with TLS support
	if config.UseTLS {
		return s.sendWithTLS(addr, auth, config.FromEmail, to, msg)
	}

	// Send without TLS
	return smtp.SendMail(addr, auth, config.FromEmail, []string{to}, []byte(msg))
}

// sendWithTLS sends email with TLS connection
func (s *EmailService) sendWithTLS(addr string, auth smtp.Auth, from, to, msg string) error {
	// Connect to server
	c, err := smtp.Dial(addr)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %w", err)
	}
	defer c.Close()

	// Start TLS
	host := strings.Split(addr, ":")[0]
	tlsConfig := &tls.Config{
		ServerName: host,
		MinVersion: tls.VersionTLS12,
	}

	if err = c.StartTLS(tlsConfig); err != nil {
		return fmt.Errorf("failed to start TLS: %w", err)
	}

	// Authenticate
	if auth != nil {
		if err = c.Auth(auth); err != nil {
			return fmt.Errorf("failed to authenticate: %w", err)
		}
	}

	// Send email
	if err = c.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %w", err)
	}

	if err = c.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %w", err)
	}

	w, err := c.Data()
	if err != nil {
		return fmt.Errorf("failed to get data writer: %w", err)
	}

	_, err = w.Write([]byte(msg))
	if err != nil {
		return fmt.Errorf("failed to write message: %w", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return c.Quit()
}

// SendPasswordResetEmail sends a password reset email
func (s *EmailService) SendPasswordResetEmail(to, username, resetToken string) error {
	subject := "Password Reset Request - Docode WAF"

	// Build email body
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; color: #333; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background: #1e40af; color: white; padding: 20px; text-align: center; }
        .content { background: #f9fafb; padding: 30px; border-radius: 8px; margin-top: 20px; }
        .button { display: inline-block; background: #3b82f6; color: white; padding: 12px 30px; 
                  text-decoration: none; border-radius: 5px; margin: 20px 0; }
        .footer { margin-top: 30px; text-align: center; font-size: 12px; color: #6b7280; }
        .warning { background: #fef3c7; border-left: 4px solid #f59e0b; padding: 12px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Password Reset Request</h1>
        </div>
        <div class="content">
            <p>Hello <strong>%s</strong>,</p>
            
            <p>We received a request to reset your password for your Docode WAF account.</p>
            
            <p>Your password reset token is:</p>
            
            <div style="background: white; padding: 15px; border-radius: 5px; font-family: monospace; font-size: 18px; text-align: center; letter-spacing: 2px; margin: 20px 0;">
                <strong>%s</strong>
            </div>
            
            <p>Use this token on the password reset page to create a new password.</p>
            
            <div class="warning">
                <strong>⚠️ Security Notice:</strong> This token will expire in 1 hour. 
                If you didn't request this password reset, please ignore this email and your password will remain unchanged.
            </div>
            
            <p>For security reasons, never share this token with anyone.</p>
            
            <p>Best regards,<br>
            <strong>Docode WAF Team</strong></p>
        </div>
        <div class="footer">
            <p>This is an automated message, please do not reply to this email.</p>
            <p>&copy; 2025 Docode WAF. All rights reserved.</p>
        </div>
    </div>
</body>
</html>
`, username, resetToken)

	return s.SendEmail(to, subject, body)
}
