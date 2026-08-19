package entities

import "gorm.io/gorm"

// ProducerComponentVariant guarda, para cada Producer, qué ComponentVariant
// eligió por módulo. Un solo registro vivo por (producer, module).
type ProducerComponentVariant struct {
	gorm.Model

	ProducerID         uint `gorm:"column:producer_id;not null;uniqueIndex:idx_pcv_producer_module,where:deleted_at IS NULL"`
	ModuleID           uint `gorm:"column:module_id;not null;uniqueIndex:idx_pcv_producer_module,where:deleted_at IS NULL"`
	ComponentVariantID uint `gorm:"column:component_variant_id;not null"`

	Producer         Producer         `gorm:"foreignKey:ProducerID"`
	Module           Module           `gorm:"foreignKey:ModuleID"`
	ComponentVariant ComponentVariant `gorm:"foreignKey:ComponentVariantID"`
}
