package entities

import (
	"time"

	"gorm.io/gorm"
)

const (
	DomainTypeSubdomain = "subdomain" // ej: fiestabresh.goaccess.com.ar
	DomainTypeCustom    = "custom"    // ej: fiestabresh.com
)

// Domain es un subdominio o dominio propio vinculado a un Producer. Es la
// clave con la que el middleware de Next.js resuelve el tenant. Los dominios
// custom requieren verificación de propiedad (columna VerifiedAt).
type Domain struct {
	gorm.Model

	ProducerID uint       `gorm:"column:producer_id;not null"`
	Domain     string     `gorm:"column:domain;not null;uniqueIndex:idx_domains_domain,where:deleted_at IS NULL"`
	Type       string     `gorm:"column:type;not null"`
	VerifiedAt *time.Time `gorm:"column:verified_at"`

	Producer Producer `gorm:"foreignKey:ProducerID"`
}
