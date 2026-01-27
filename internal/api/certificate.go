package api

import (
	"net/http"

	"github.com/aleh/docode-waf/internal/models"
	"github.com/aleh/docode-waf/internal/services"
	"github.com/gin-gonic/gin"
)

type CertificateHandler struct {
	certService *services.CertificateService
	acmeService *services.ACMEService
}

func NewCertificateHandler(certService *services.CertificateService, acmeService *services.ACMEService) *CertificateHandler {
	return &CertificateHandler{
		certService: certService,
		acmeService: acmeService,
	}
}

// CreateCertificate handles certificate creation
func (h *CertificateHandler) CreateCertificate(c *gin.Context) {
	var input models.CertificateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	certificate, err := h.certService.CreateCertificate(&input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Certificate created successfully",
		"certificate": certificate,
	})
}

// GetCertificates retrieves all certificates
func (h *CertificateHandler) GetCertificates(c *gin.Context) {
	certificates, err := h.certService.GetAllCertificates()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"certificates": certificates,
		"total":        len(certificates),
	})
}

// GetCertificate retrieves a single certificate
func (h *CertificateHandler) GetCertificate(c *gin.Context) {
	id := c.Param("id")

	certificate, err := h.certService.GetCertificate(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Certificate not found"})
		return
	}

	c.JSON(http.StatusOK, certificate)
}

// UpdateCertificate updates a certificate
func (h *CertificateHandler) UpdateCertificate(c *gin.Context) {
	id := c.Param("id")

	var input models.CertificateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	certificate, err := h.certService.UpdateCertificate(id, &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Certificate updated successfully",
		"certificate": certificate,
	})
}

// DeleteCertificate deletes a certificate
func (h *CertificateHandler) DeleteCertificate(c *gin.Context) {
	id := c.Param("id")

	err := h.certService.DeleteCertificate(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Certificate deleted successfully",
	})
}

// UploadCertificate handles file upload and creates certificate
func (h *CertificateHandler) UploadCertificate(c *gin.Context) {
	name := c.PostForm("name")
	if name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Certificate name is required"})
		return
	}

	// Get certificate file
	certFile, err := c.FormFile("cert_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Certificate file is required"})
		return
	}

	// Get key file
	keyFile, err := c.FormFile("key_file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Private key file is required"})
		return
	}

	// Read certificate content
	certReader, err := certFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read certificate file"})
		return
	}
	defer certReader.Close()

	certContent := make([]byte, certFile.Size)
	if _, err := certReader.Read(certContent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read certificate content"})
		return
	}

	// Read key content
	keyReader, err := keyFile.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read key file"})
		return
	}
	defer keyReader.Close()

	keyContent := make([]byte, keyFile.Size)
	if _, err := keyReader.Read(keyContent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read key content"})
		return
	}

	// Create certificate input
	input := &models.CertificateInput{
		Name:        name,
		CertContent: string(certContent),
		KeyContent:  string(keyContent),
	}

	certificate, err := h.certService.CreateCertificate(input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Save certificate files to filesystem
	if err := h.certService.SaveCertificateFiles(certificate.ID, certContent, keyContent); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save certificate files"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Certificate uploaded successfully",
		"certificate": certificate,
	})
}

// GenerateCertificate handles certificate generation via Let's Encrypt
func (h *CertificateHandler) GenerateCertificate(c *gin.Context) {
	var input struct {
		Domain             string `json:"domain" binding:"required"`
		Email              string `json:"email" binding:"required,email"`
		IsWildcard         bool   `json:"is_wildcard"`
		DNSProvider        string `json:"dns_provider"`
		CloudflareAPIToken string `json:"cloudflare_api_token"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Prepare credentials map
	credentials := make(map[string]string)
	if input.DNSProvider == "cloudflare" {
		credentials["cloudflare_api_token"] = input.CloudflareAPIToken
	}

	// Adjust domain for wildcard
	targetDomain := input.Domain
	if input.IsWildcard {
		// Ensure domain starts with *.
		if len(targetDomain) < 2 || targetDomain[:2] != "*." {
			targetDomain = "*." + targetDomain
		}
		// Wildcard requires DNS-01
		if input.DNSProvider == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Wildcard certificates require a DNS provider (e.g., Cloudflare)"})
			return
		}
	}

	// Request certificate from Let's Encrypt
	certs, err := h.acmeService.ObtainCertificate(targetDomain, input.Email, input.DNSProvider, credentials)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate certificate: " + err.Error()})
		return
	}

	// Create certificate input for DB
	certName := targetDomain
	if input.IsWildcard {
		certName = "Wildcard " + input.Domain
	}

	certInput := &models.CertificateInput{
		Name:        certName,
		CertContent: string(certs.Certificate),
		KeyContent:  string(certs.PrivateKey),
	}

	// Save to DB
	certificate, err := h.certService.CreateCertificate(certInput)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save generated certificate: " + err.Error()})
		return
	}

	// Save files to filesystem
	if err := h.certService.SaveCertificateFiles(certificate.ID, []byte(certInput.CertContent), []byte(certInput.KeyContent)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save certificate files: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "Certificate generated successfully",
		"certificate": certificate,
	})
}

// UpdateCertificateStatuses manually triggers status update for all certificates
func (h *CertificateHandler) UpdateCertificateStatuses(c *gin.Context) {
	err := h.certService.UpdateCertificateStatuses()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Certificate statuses updated successfully",
	})
}
