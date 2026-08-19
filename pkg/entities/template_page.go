package entities

import (
	"encoding/json"

	"gorm.io/gorm"
)

// TemplatePage es la estructura Puck por defecto para un tipo de página
// dentro de una Template. Al aplicar la template a un Producer, se copia
// como PageTemplate editable.
type TemplatePage struct {
	gorm.Model

	TemplateID      uint            `gorm:"column:template_id;not null;uniqueIndex:idx_template_pages_pair,where:deleted_at IS NULL"`
	PageType        string          `gorm:"column:page_type;not null;uniqueIndex:idx_template_pages_pair,where:deleted_at IS NULL"`
	PuckJSONDefault json.RawMessage `gorm:"column:puck_json_default;type:jsonb"`

	Template Template `gorm:"foreignKey:TemplateID"`
}
