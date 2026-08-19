package entities

import "gorm.io/gorm"

const (
	ProducerStatusActive    = "active"
	ProducerStatusSuspended = "suspended"
)

// Producer representa un tenant (productora/organizador) administrado por el
// backoffice. Es la unidad base sobre la que se configuran módulos, variantes,
// design tokens, páginas Puck, dominios y comisiones.
//
// Dueño de escritura: este servicio (ticketing-platform). ticketing-core solo
// lo lee para resolver contexto de la tienda pública.
type Producer struct {
	gorm.Model

	Name         string `gorm:"column:name;not null"`
	Slug         string `gorm:"column:slug;uniqueIndex:idx_producers_slug,where:deleted_at IS NULL;not null"`
	ContactEmail string `gorm:"column:contact_email"`
	Status       string `gorm:"column:status;default:'active'"`

	TemplateID *uint     `gorm:"column:template_id"`
	Template   *Template `gorm:"foreignKey:TemplateID"`

	PlanID *uint `gorm:"column:plan_id"`
	Plan   *Plan `gorm:"foreignKey:PlanID"`
}
