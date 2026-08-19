package entities

import "gorm.io/gorm"

// Plan representa un plan comercial (etiqueta) que se le asigna a un Producer.
// Es puramente descriptivo — la habilitación real de módulos vive en
// ProducerModule. Un cambio de plan no altera ProducerModule automáticamente.
type Plan struct {
	gorm.Model

	Code        string `gorm:"column:code;uniqueIndex:idx_plans_code,where:deleted_at IS NULL;not null"`
	Name        string `gorm:"column:name;not null"`
	Description string `gorm:"column:description"`
	IsActive    bool   `gorm:"column:is_active;default:true"`
}
