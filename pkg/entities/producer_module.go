package entities

import "gorm.io/gorm"

// ProducerModule es la FUENTE ÚNICA DE VERDAD sobre qué módulos ve un Producer.
// El plan comercial es solo etiqueta; lo que efectivamente se renderiza en la
// tienda pública se decide leyendo esta tabla.
type ProducerModule struct {
	gorm.Model

	ProducerID uint `gorm:"column:producer_id;not null;uniqueIndex:idx_producer_modules_pair,where:deleted_at IS NULL"`
	ModuleID   uint `gorm:"column:module_id;not null;uniqueIndex:idx_producer_modules_pair,where:deleted_at IS NULL"`
	Enabled    bool `gorm:"column:enabled;default:true"`

	Producer Producer `gorm:"foreignKey:ProducerID"`
	Module   Module   `gorm:"foreignKey:ModuleID"`
}
