package pagetemplatedtos

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
)

var (
	ErrPageNotFound    = errors.New("página no encontrada")
	ErrPageTypeInvalid = errors.New("el tipo de página no es válido")
	ErrPuckJSONBroken  = errors.New("el puck_json enviado no es JSON válido")
	ErrPuckJSONEmpty   = errors.New("puck_json es requerido")
)

type RequestSavePage struct {
	PuckJSON json.RawMessage `json:"puck_json"`
}

// Validate garantiza que el JSON de Puck sea sintácticamente válido antes
// de persistirlo. Guardar JSON roto haría que la página del tenant no
// renderice en la tienda pública — este es el guardarraíl mínimo.
// La validación estructural contra config.components queda para una fase
// posterior (ver §8 del spec).
func (r RequestSavePage) Validate() error {
	if len(r.PuckJSON) == 0 {
		return ErrPuckJSONEmpty
	}
	if !json.Valid(r.PuckJSON) {
		return ErrPuckJSONBroken
	}
	return nil
}

type ResponsePageTemplate struct {
	ID         uint            `json:"id"`
	ProducerID uint            `json:"producer_id"`
	PageType   string          `json:"page_type"`
	PuckJSON   json.RawMessage `json:"puck_json,omitempty"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

func (r *ResponsePageTemplate) FromEntity(e entities.PageTemplate) {
	r.ID = e.ID
	r.ProducerID = e.ProducerID
	r.PageType = e.PageType
	r.PuckJSON = e.PuckJSON
	r.UpdatedAt = e.UpdatedAt
}

type ResponsePageTemplates struct {
	Pages []ResponsePageTemplate `json:"pages"`
}

func (r *ResponsePageTemplates) FromEntities(list []entities.PageTemplate) {
	r.Pages = make([]ResponsePageTemplate, len(list))
	for i, e := range list {
		r.Pages[i].FromEntity(e)
	}
}

func IsValidPageType(pt string) bool {
	switch pt {
	case entities.PageTypeHome,
		entities.PageTypeEventDetail,
		entities.PageTypeCheckout,
		entities.PageTypeGallery,
		entities.PageTypeFAQ,
		entities.PageTypeContact:
		return true
	}
	return false
}
