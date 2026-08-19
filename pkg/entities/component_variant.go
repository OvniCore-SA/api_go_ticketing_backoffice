package entities

import "gorm.io/gorm"

// ComponentVariant es una disposición visual concreta para un módulo
// (por ej. grid, carrusel, masonry para el módulo Gallery). Cada Producer
// elige, por módulo, una variante en producer_component_variants.
type ComponentVariant struct {
	gorm.Model

	ModuleID    uint   `gorm:"column:module_id;not null;uniqueIndex:idx_component_variants_module_code,where:deleted_at IS NULL"`
	Code        string `gorm:"column:code;not null;uniqueIndex:idx_component_variants_module_code,where:deleted_at IS NULL"`
	Name        string `gorm:"column:name;not null"`
	Description string `gorm:"column:description"`
	PreviewURL  string `gorm:"column:preview_url"`

	Module Module `gorm:"foreignKey:ModuleID"`
}
