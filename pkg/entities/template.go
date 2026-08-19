package entities

import (
	"encoding/json"

	"gorm.io/gorm"
)

// Template es un preset de fábrica (colores, fuentes, set de módulos, páginas
// Puck iniciales) que se puede aplicar al crear un Producer. El seed copia
// TemplateModule → ProducerModule, TemplatePage → PageTemplate y default_*
// → DesignTokens dentro de una transacción.
type Template struct {
	gorm.Model

	Code           string          `gorm:"column:code;uniqueIndex:idx_templates_code,where:deleted_at IS NULL;not null"`
	Name           string          `gorm:"column:name;not null"`
	Description    string          `gorm:"column:description"`
	PreviewURL     string          `gorm:"column:preview_url"`
	DefaultColors  json.RawMessage `gorm:"column:default_colors;type:jsonb"`
	DefaultFonts   json.RawMessage `gorm:"column:default_fonts;type:jsonb"`
	DefaultRadius  json.RawMessage `gorm:"column:default_radius;type:jsonb"`
	DefaultShadows json.RawMessage `gorm:"column:default_shadows;type:jsonb"`

	Modules []TemplateModule `gorm:"foreignKey:TemplateID"`
	Pages   []TemplatePage   `gorm:"foreignKey:TemplateID"`
}
