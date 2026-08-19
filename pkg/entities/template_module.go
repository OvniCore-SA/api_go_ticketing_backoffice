package entities

import "gorm.io/gorm"

// TemplateModule es el set de módulos que trae de fábrica una Template.
// Al aplicar la template a un Producer, cada entrada se copia como
// ProducerModule con enabled=true.
type TemplateModule struct {
	gorm.Model

	TemplateID uint `gorm:"column:template_id;not null;uniqueIndex:idx_template_modules_pair,where:deleted_at IS NULL"`
	ModuleID   uint `gorm:"column:module_id;not null;uniqueIndex:idx_template_modules_pair,where:deleted_at IS NULL"`

	Template Template `gorm:"foreignKey:TemplateID"`
	Module   Module   `gorm:"foreignKey:ModuleID"`
}
