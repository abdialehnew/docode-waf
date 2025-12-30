package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/aleh/docode-waf/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

// AuthService handles authentication operations
type AuthService struct {
	db        *sqlx.DB
	jwtSecret []byte
}

// NewAuthService creates a new auth service
func NewAuthService(db *sqlx.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

// Register creates a new admin user
func (s *AuthService) Register(username, email, password, fullName string) (*models.Admin, error) {
	// Check if user already exists
	var exists bool
	err := s.db.Get(&exists, "SELECT EXISTS(SELECT 1 FROM admins WHERE username = $1 OR email = $2)", username, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Insert admin
	query := `
		INSERT INTO admins (id, username, email, password_hash, full_name, role, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, $4, 'super_admin', CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		RETURNING id, username, email, full_name, role, is_active, created_at, updated_at
	`

	admin := &models.Admin{}
	err = s.db.QueryRowx(query, username, email, string(hashedPassword), fullName).StructScan(admin)
	if err != nil {
		return nil, err
	}

	return admin, nil
}

// Login authenticates a user and returns a JWT token
func (s *AuthService) Login(username, password string) (string, *models.Admin, error) {
	// Get admin by username or email
	admin := &models.Admin{}
	query := `
		SELECT id, username, email, password_hash, full_name, role, is_active, last_login, created_at, updated_at
		FROM admins
		WHERE (username = $1 OR email = $1) AND is_active = true
	`
	err := s.db.QueryRowx(query, username).StructScan(admin)
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// Verify password
	err = bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password))
	if err != nil {
		return "", nil, ErrInvalidCredentials
	}

	// Update last login
	s.db.Exec("UPDATE admins SET last_login = CURRENT_TIMESTAMP WHERE id = $1", admin.ID)

	// Generate JWT token
	token, err := s.generateToken(admin)
	if err != nil {
		return "", nil, err
	}

	return token, admin, nil
}

// RequestPasswordReset generates a reset token and stores it
func (s *AuthService) RequestPasswordReset(email string) (string, string, error) {
	// Check if admin exists and get username
	var admin struct {
		ID       string `db:"id"`
		Username string `db:"username"`
	}
	err := s.db.Get(&admin, "SELECT id, username FROM admins WHERE email = $1 AND is_active = true", email)
	if err != nil {
		return "", "", ErrUserNotFound
	}

	// Generate reset token
	token := generateRandomToken(32)
	expiry := time.Now().Add(1 * time.Hour)

	// Store token
	query := `
		UPDATE admins 
		SET reset_token = $1, reset_token_expiry = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`
	_, err = s.db.Exec(query, token, expiry, admin.ID)
	if err != nil {
		return "", "", err
	}

	return admin.Username, token, nil
}

// ResetPassword resets password using reset token
func (s *AuthService) ResetPassword(token, newPassword string) error {
	// Find admin with valid token
	admin := &models.Admin{}
	query := `
		SELECT id, reset_token_expiry
		FROM admins
		WHERE reset_token = $1 AND is_active = true
	`
	err := s.db.QueryRowx(query, token).StructScan(admin)
	if err != nil {
		return ErrInvalidToken
	}

	// Check token expiry
	if admin.ResetTokenExpiry == nil || time.Now().After(*admin.ResetTokenExpiry) {
		return ErrInvalidToken
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password and clear reset token
	updateQuery := `
		UPDATE admins 
		SET password_hash = $1, reset_token = NULL, reset_token_expiry = NULL, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	_, err = s.db.Exec(updateQuery, string(hashedPassword), admin.ID)
	return err
}

// ChangePassword changes user password (requires old password)
func (s *AuthService) ChangePassword(adminID, oldPassword, newPassword string) error {
	// Get current password hash
	var currentHash string
	err := s.db.Get(&currentHash, "SELECT password_hash FROM admins WHERE id = $1", adminID)
	if err != nil {
		return ErrUserNotFound
	}

	// Verify old password
	err = bcrypt.CompareHashAndPassword([]byte(currentHash), []byte(oldPassword))
	if err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Update password
	_, err = s.db.Exec("UPDATE admins SET password_hash = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2", string(hashedPassword), adminID)
	return err
}

// ValidateToken validates a JWT token and returns the admin
func (s *AuthService) ValidateToken(tokenString string) (*models.Admin, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	adminID, ok := claims["admin_id"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Get admin from database
	admin := &models.Admin{}
	query := `
		SELECT id, username, email, full_name, role, is_active, created_at, updated_at
		FROM admins
		WHERE id = $1 AND is_active = true
	`
	err = s.db.QueryRowx(query, adminID).StructScan(admin)
	if err != nil {
		return nil, ErrUserNotFound
	}

	return admin, nil
}

func (s *AuthService) generateToken(admin *models.Admin) (string, error) {
	claims := jwt.MapClaims{
		"admin_id": admin.ID,
		"username": admin.Username,
		"email":    admin.Email,
		"role":     admin.Role,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.jwtSecret)
}

func generateRandomToken(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return hex.EncodeToString(b)
}
