package entities

import "gorm.io/gorm"

// Códigos de módulos del catálogo global. Un módulo nuevo siempre implica
// código nuevo en el frontend (bloque de Puck), así que agregarlo acá y en
// el seed es suficiente — no hay UI para crear módulos.
const (
	ModuleCodeTicketingCore = "ticketing_core"
	ModuleCodeCartelera     = "cartelera"
	ModuleCodeGallery       = "gallery"
	ModuleCodeFAQ           = "faq"
	ModuleCodeMap           = "map"
	ModuleCodePromotions    = "promotions"
	ModuleCodePOS           = "pos"
)

// Module es un ítem del catálogo global de funcionalidades que un Producer
// puede tener habilitado. Los módulos marcados con IsCore no se pueden
// deshabilitar en ningún Producer (regla que se aplica en producer_modules).
type Module struct {
	gorm.Model

	Code        string `gorm:"column:code;uniqueIndex:idx_modules_code,where:deleted_at IS NULL;not null"`
	Name        string `gorm:"column:name;not null"`
	Description string `gorm:"column:description"`
	Category    string `gorm:"column:category"`
	IsCore      bool   `gorm:"column:is_core;default:false"`

	Variants []ComponentVariant `gorm:"foreignKey:ModuleID"`
}
