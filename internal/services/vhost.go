package services

import (
	"github.com/aleh/docode-waf/internal/models"
	"github.com/jmoiron/sqlx"
)
type VHostService struct {
	db *sqlx.DB
}
func NewVHostService(db *sqlx.DB) *VHostService {
	return &VHostService{db: db}
func (s *VHostService) GetAll() ([]*models.VHost, error) {
	var vhosts []*models.VHost
	err := s.db.Select(&vhosts, "SELECT * FROM vhosts WHERE enabled = true")
	return vhosts, err
func (s *VHostService) GetByID(id string) (*models.VHost, error) {
	var vhost models.VHost
	err := s.db.Get(&vhost, "SELECT * FROM vhosts WHERE id = $1", id)
	return &vhost, err
func (s *VHostService) Create(vhost *models.VHost) error {
	query := `INSERT INTO vhosts (name, domain, backend_url, ssl_enabled, ssl_cert_path, ssl_key_path, enabled)
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, created_at, updated_at`
	
	return s.db.QueryRowx(query,
		vhost.Name,
		vhost.Domain,
		vhost.BackendURL,
		vhost.SSLEnabled,
		vhost.SSLCertPath,
}	return locations, err		vhostID)		"SELECT * FROM vhost_locations WHERE vhost_id = $1 AND enabled = true", 	err := s.db.Select(&locations, 	var locations []*models.VHostLocationfunc (s *VHostService) GetLocations(vhostID string) ([]*models.VHostLocation, error) {}	return err	_, err := s.db.Exec("DELETE FROM vhosts WHERE id = $1", id)func (s *VHostService) Delete(id string) error {}	return err	)		vhost.ID,		vhost.Enabled,		vhost.SSLKeyPath,		vhost.SSLCertPath,		vhost.SSLEnabled,		vhost.BackendURL,		vhost.Domain,		vhost.Name,	_, err := s.db.Exec(query,				  WHERE id=$8`			  SET name=$1, domain=$2, backend_url=$3, ssl_enabled=$4, ssl_cert_path=$5, ssl_key_path=$6, enabled=$7	query := `UPDATE vhosts func (s *VHostService) Update(vhost *models.VHost) error {}	).Scan(&vhost.ID, &vhost.CreatedAt, &vhost.UpdatedAt)		vhost.Enabled,		vhost.SSLKeyPath,
