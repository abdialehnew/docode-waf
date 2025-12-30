package models

import (
	"time"
)

type Certificate struct {
	ID          string    `db:"id" json:"id"`
	Name        string    `db:"name" json:"name"`
	CertContent string    `db:"cert_content" json:"cert_content"`
	KeyContent  string    `db:"key_content" json:"key_content"`
	CommonName  string    `db:"common_name" json:"common_name"`
	Issuer      string    `db:"issuer" json:"issuer"`
	ValidFrom   time.Time `db:"valid_from" json:"valid_from"`
	ValidTo     time.Time `db:"valid_to" json:"valid_to"`
	Status      string    `db:"status" json:"status"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}

type CertificateInput struct {
	Name        string `json:"name" binding:"required,min=3"`
	CertContent string `json:"cert_content" binding:"required"`
	KeyContent  string `json:"key_content" binding:"required"`
}
