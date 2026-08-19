package entities

import (
	"encoding/json"

	"gorm.io/gorm"
)

// DesignTokens contiene el tema visual real de un Producer (paleta, fuentes,
// radios, sombras). Los cuatro campos son JSON libre; el shape lo decide el
// frontend y este servicio no lo interpreta — solo lo guarda y lo devuelve.
// Un solo registro vivo por producer.
type DesignTokens struct {
	gorm.Model

	ProducerID uint            `gorm:"column:producer_id;not null;uniqueIndex:idx_design_tokens_producer,where:deleted_at IS NULL"`
	Colors     json.RawMessage `gorm:"column:colors;type:jsonb"`
	Fonts      json.RawMessage `gorm:"column:fonts;type:jsonb"`
	Radius     json.RawMessage `gorm:"column:radius;type:jsonb"`
	Shadows    json.RawMessage `gorm:"column:shadows;type:jsonb"`

	Producer Producer `gorm:"foreignKey:ProducerID"`
}
