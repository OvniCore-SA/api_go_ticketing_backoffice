package entities

import (
	"encoding/json"

	"gorm.io/gorm"
)

// Tipos de página soportados en el sistema. Cada Producer puede tener a lo
// sumo un PageTemplate vivo por (producer, page_type). Agregar nuevos tipos
// requiere que el frontend sepa renderizarlos con Puck.
const (
	PageTypeHome        = "home"
	PageTypeEventDetail = "event_detail"
	PageTypeCheckout    = "checkout"
	PageTypeGallery     = "gallery"
	PageTypeFAQ         = "faq"
	PageTypeContact     = "contact"
)

// PageTemplate es la estructura Puck real y editable de una página concreta
// de un Producer. Es el destino del `onPublish` del editor y la fuente que
// consume ticketing-core cuando resuelve la tienda pública.
type PageTemplate struct {
	gorm.Model

	ProducerID uint            `gorm:"column:producer_id;not null;uniqueIndex:idx_page_templates_pair,where:deleted_at IS NULL"`
	PageType   string          `gorm:"column:page_type;not null;uniqueIndex:idx_page_templates_pair,where:deleted_at IS NULL"`
	PuckJSON   json.RawMessage `gorm:"column:puck_json;type:jsonb"`

	Producer Producer `gorm:"foreignKey:ProducerID"`
}
