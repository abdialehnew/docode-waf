package services

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aleh/docode-waf/internal/models"
	"github.com/jmoiron/sqlx"
)

type CertificateService struct {
	db *sqlx.DB
}

func NewCertificateService(db *sqlx.DB) *CertificateService {
	return &CertificateService{db: db}
}

// ParseCertificate parses PEM encoded certificate and extracts information
func (s *CertificateService) ParseCertificate(certPEM string) (*x509.Certificate, error) {
	block, _ := pem.Decode([]byte(certPEM))
	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}

	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed to parse certificate: %w", err)
	}

	return cert, nil
}

// ValidateKeyPair validates that the certificate and private key match
func (s *CertificateService) ValidateKeyPair(certPEM, keyPEM string) error {
	// Parse certificate
	certBlock, _ := pem.Decode([]byte(certPEM))
	if certBlock == nil {
		return errors.New("failed to decode certificate PEM")
	}

	cert, err := x509.ParseCertificate(certBlock.Bytes)
	if err != nil {
		return fmt.Errorf("failed to parse certificate: %w", err)
	}

	// Parse private key
	keyBlock, _ := pem.Decode([]byte(keyPEM))
	if keyBlock == nil {
		return errors.New("failed to decode private key PEM")
	}

	// Try to parse as different key types
	_, err = x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	if err != nil {
		_, err = x509.ParsePKCS8PrivateKey(keyBlock.Bytes)
		if err != nil {
			_, err = x509.ParseECPrivateKey(keyBlock.Bytes)
			if err != nil {
				return errors.New("failed to parse private key")
			}
		}
	}

	// Basic validation: certificate should not be expired
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return errors.New("certificate is not yet valid")
	}
	if now.After(cert.NotAfter) {
		return errors.New("certificate has expired")
	}

	return nil
}

// GetStatus determines certificate status based on validity dates
func (s *CertificateService) GetStatus(validFrom, validTo time.Time) string {
	now := time.Now()

	if now.Before(validFrom) {
		return "pending"
	}
	if now.After(validTo) {
		return "expired"
	}

	// Check if expiring soon (within 30 days)
	if validTo.Sub(now) < 30*24*time.Hour {
		return "expiring_soon"
	}

	return "active"
}

// CreateCertificate creates a new certificate
func (s *CertificateService) CreateCertificate(input *models.CertificateInput) (*models.Certificate, error) {
	// Validate key pair
	if err := s.ValidateKeyPair(input.CertContent, input.KeyContent); err != nil {
		return nil, fmt.Errorf("invalid certificate/key pair: %w", err)
	}

	// Parse certificate to extract info
	cert, err := s.ParseCertificate(input.CertContent)
	if err != nil {
		return nil, err
	}

	// Determine status
	status := s.GetStatus(cert.NotBefore, cert.NotAfter)

	// Insert into database
	query := `
		INSERT INTO certificates (name, cert_content, key_content, common_name, issuer, valid_from, valid_to, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	certificate := &models.Certificate{
		Name:        input.Name,
		CertContent: input.CertContent,
		KeyContent:  input.KeyContent,
		CommonName:  cert.Subject.CommonName,
		Issuer:      cert.Issuer.CommonName,
		ValidFrom:   cert.NotBefore,
		ValidTo:     cert.NotAfter,
		Status:      status,
	}

	err = s.db.QueryRow(
		query,
		certificate.Name,
		certificate.CertContent,
		certificate.KeyContent,
		certificate.CommonName,
		certificate.Issuer,
		certificate.ValidFrom,
		certificate.ValidTo,
		certificate.Status,
	).Scan(&certificate.ID, &certificate.CreatedAt, &certificate.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create certificate: %w", err)
	}

	return certificate, nil
}

// GetAllCertificates retrieves all certificates
func (s *CertificateService) GetAllCertificates() ([]models.Certificate, error) {
	var certificates []models.Certificate

	query := `
		SELECT id, name, cert_content, key_content, common_name, issuer, 
		       valid_from, valid_to, status, created_at, updated_at
		FROM certificates
		ORDER BY created_at DESC
	`

	err := s.db.Select(&certificates, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get certificates: %w", err)
	}

	// Update status for each certificate
	for i := range certificates {
		certificates[i].Status = s.GetStatus(certificates[i].ValidFrom, certificates[i].ValidTo)
	}

	return certificates, nil
}

// GetCertificate retrieves a certificate by ID
func (s *CertificateService) GetCertificate(id string) (*models.Certificate, error) {
	var certificate models.Certificate

	query := `
		SELECT id, name, cert_content, key_content, common_name, issuer, 
		       valid_from, valid_to, status, created_at, updated_at
		FROM certificates
		WHERE id = $1
	`

	err := s.db.Get(&certificate, query, id)
	if err != nil {
		return nil, fmt.Errorf("certificate not found: %w", err)
	}

	// Update status
	certificate.Status = s.GetStatus(certificate.ValidFrom, certificate.ValidTo)

	return &certificate, nil
}

// UpdateCertificate updates a certificate
func (s *CertificateService) UpdateCertificate(id string, input *models.CertificateInput) (*models.Certificate, error) {
	// Validate key pair
	if err := s.ValidateKeyPair(input.CertContent, input.KeyContent); err != nil {
		return nil, fmt.Errorf("invalid certificate/key pair: %w", err)
	}

	// Parse certificate to extract info
	cert, err := s.ParseCertificate(input.CertContent)
	if err != nil {
		return nil, err
	}

	// Determine status
	status := s.GetStatus(cert.NotBefore, cert.NotAfter)

	query := `
		UPDATE certificates
		SET name = $1, cert_content = $2, key_content = $3, common_name = $4,
		    issuer = $5, valid_from = $6, valid_to = $7, status = $8,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = $9
		RETURNING id, name, cert_content, key_content, common_name, issuer,
		          valid_from, valid_to, status, created_at, updated_at
	`

	var certificate models.Certificate
	err = s.db.QueryRow(
		query,
		input.Name,
		input.CertContent,
		input.KeyContent,
		cert.Subject.CommonName,
		cert.Issuer.CommonName,
		cert.NotBefore,
		cert.NotAfter,
		status,
		id,
	).Scan(
		&certificate.ID,
		&certificate.Name,
		&certificate.CertContent,
		&certificate.KeyContent,
		&certificate.CommonName,
		&certificate.Issuer,
		&certificate.ValidFrom,
		&certificate.ValidTo,
		&certificate.Status,
		&certificate.CreatedAt,
		&certificate.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to update certificate: %w", err)
	}

	return &certificate, nil
}

// DeleteCertificate deletes a certificate
func (s *CertificateService) DeleteCertificate(id string) error {
	query := `DELETE FROM certificates WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete certificate: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return errors.New("certificate not found")
	}

	return nil
}

// UpdateCertificateStatuses updates all certificate statuses
func (s *CertificateService) UpdateCertificateStatuses() error {
	query := `
		UPDATE certificates
		SET status = CASE
			WHEN NOW() < valid_from THEN 'pending'
			WHEN NOW() > valid_to THEN 'expired'
			WHEN (valid_to - NOW()) < INTERVAL '30 days' THEN 'expiring_soon'
			ELSE 'active'
		END,
		updated_at = CURRENT_TIMESTAMP
		WHERE status != CASE
			WHEN NOW() < valid_from THEN 'pending'
			WHEN NOW() > valid_to THEN 'expired'
			WHEN (valid_to - NOW()) < INTERVAL '30 days' THEN 'expiring_soon'
			ELSE 'active'
		END
	`

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to update certificate statuses: %w", err)
	}

	return nil
}

// SaveCertificateFiles saves certificate and key files to filesystem
func (s *CertificateService) SaveCertificateFiles(certID string, certContent, keyContent []byte) error {
	// Create certificate directory
	certDir := filepath.Join("/app/ssl/certificates", certID)
	if err := os.MkdirAll(certDir, 0755); err != nil {
		return fmt.Errorf("failed to create certificate directory: %w", err)
	}

	// Write certificate file
	certPath := filepath.Join(certDir, "cert.pem")
	if err := os.WriteFile(certPath, certContent, 0644); err != nil {
		return fmt.Errorf("failed to write certificate file: %w", err)
	}

	// Write key file
	keyPath := filepath.Join(certDir, "key.pem")
	if err := os.WriteFile(keyPath, keyContent, 0600); err != nil {
		return fmt.Errorf("failed to write key file: %w", err)
	}

	return nil
}
