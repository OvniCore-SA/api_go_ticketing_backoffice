package producervariantdtos

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
)

var (
	ErrVariantRequired   = errors.New("se requiere component_variant_id")
	ErrVariantNotInModule = errors.New("la variante no pertenece al módulo indicado")
	ErrProducerVariantNotFound = errors.New("no hay variante configurada para este módulo")
)

type RequestAssignVariant struct {
	ComponentVariantID uint `json:"component_variant_id"`
}

func (r RequestAssignVariant) Validate() error {
	if r.ComponentVariantID == 0 {
		return ErrVariantRequired
	}
	return nil
}

type ResponseProducerComponentVariant struct {
	ID                 uint   `json:"id"`
	ProducerID         uint   `json:"producer_id"`
	ModuleID           uint   `json:"module_id"`
	ComponentVariantID uint   `json:"component_variant_id"`
	VariantCode        string `json:"variant_code,omitempty"`
}

func (r *ResponseProducerComponentVariant) FromEntity(e entities.ProducerComponentVariant) {
	r.ID = e.ID
	r.ProducerID = e.ProducerID
	r.ModuleID = e.ModuleID
	r.ComponentVariantID = e.ComponentVariantID
	r.VariantCode = e.ComponentVariant.Code
}

type ResponseProducerComponentVariants struct {
	Items []ResponseProducerComponentVariant `json:"items"`
}

func (r *ResponseProducerComponentVariants) FromEntities(list []entities.ProducerComponentVariant) {
	r.Items = make([]ResponseProducerComponentVariant, len(list))
	for i, e := range list {
		r.Items[i].FromEntity(e)
	}
}
