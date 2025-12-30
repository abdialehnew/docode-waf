package services

import (
	"fmt"

	"github.com/aleh/docode-waf/internal/models"
	"github.com/jmoiron/sqlx"
)

// VHostService handles virtual host operations
type VHostService struct {
	db *sqlx.DB
}

// NewVHostService creates a new vhost service
func NewVHostService(db *sqlx.DB) *VHostService {
	return &VHostService{db: db}
}

// GetVHostByDomain retrieves a virtual host by domain name
func (s *VHostService) GetVHostByDomain(domain string) (map[string]interface{}, error) {
	vhost := make(map[string]interface{})

	query := `
		SELECT id, name, domain, backend_url, ssl_enabled, 
		       ssl_cert_path, ssl_key_path, enabled
		FROM vhosts 
		WHERE domain = $1 AND enabled = true
	`

	rows, err := s.db.Queryx(query, domain)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		err := rows.MapScan(vhost)
		if err != nil {
			return nil, err
		}
		return vhost, nil
	}

	return nil, fmt.Errorf("vhost not found for domain: %s", domain)
}

// GetVHostLocations retrieves all locations for a virtual host
func (s *VHostService) GetVHostLocations(vhostID string) ([]map[string]interface{}, error) {
	var locations []map[string]interface{}

	query := `
		SELECT id, path, backend_url, enabled
		FROM vhost_locations 
		WHERE vhost_id = $1 AND enabled = true
		ORDER BY length(path) DESC
	`

	rows, err := s.db.Queryx(query, vhostID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		location := make(map[string]interface{})
		err := rows.MapScan(location)
		if err != nil {
			continue
		}
		locations = append(locations, location)
	}

	return locations, nil
}

// ListVHosts retrieves all enabled virtual hosts
func (s *VHostService) ListVHosts() ([]*models.VHost, error) {
	var vhosts []*models.VHost

	query := `
		SELECT v.id::text as id, v.name, v.domain, v.backend_url, v.ssl_enabled, 
		       v.ssl_certificate_id::text as ssl_certificate_id, 
		       v.ssl_cert_path, v.ssl_key_path, v.enabled, v.created_at, v.updated_at
		FROM vhosts v
		WHERE v.enabled = true
		ORDER BY v.created_at DESC
	`

	err := s.db.Select(&vhosts, query)
	if err != nil {
		return nil, err
	}

	return vhosts, nil
}

// GetVHostByID retrieves a virtual host by ID
func (s *VHostService) GetVHostByID(id string) (*models.VHost, error) {
	var vhost models.VHost

	query := `
		SELECT v.id::text as id, v.name, v.domain, v.backend_url, v.ssl_enabled, 
		       v.ssl_certificate_id::text as ssl_certificate_id,
		       v.ssl_cert_path, v.ssl_key_path, v.enabled, v.created_at, v.updated_at
		FROM vhosts v
		WHERE v.id = $1
	`

	err := s.db.Get(&vhost, query, id)
	if err != nil {
		return nil, fmt.Errorf("vhost not found: %w", err)
	}

	return &vhost, nil
}
