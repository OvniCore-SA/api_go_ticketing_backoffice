package producermoduledtos

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
)

var (
	ErrProducerModuleNotFound = errors.New("no hay configuración de este módulo para el producer")
	ErrCoreModuleDisallowed   = errors.New("no se puede deshabilitar un módulo core")
	ErrEnabledRequired        = errors.New("se requiere el campo enabled")
)

type RequestToggleModule struct {
	Enabled *bool `json:"enabled"`
}

func (r RequestToggleModule) Validate() error {
	if r.Enabled == nil {
		return ErrEnabledRequired
	}
	return nil
}

type ResponseProducerModule struct {
	ID         uint   `json:"id"`
	ProducerID uint   `json:"producer_id"`
	ModuleID   uint   `json:"module_id"`
	ModuleCode string `json:"module_code,omitempty"`
	Enabled    bool   `json:"enabled"`
}

func (r *ResponseProducerModule) FromEntity(e entities.ProducerModule) {
	r.ID = e.ID
	r.ProducerID = e.ProducerID
	r.ModuleID = e.ModuleID
	r.ModuleCode = e.Module.Code
	r.Enabled = e.Enabled
}

type ResponseProducerModules struct {
	Modules []ResponseProducerModule `json:"modules"`
}

func (r *ResponseProducerModules) FromEntities(list []entities.ProducerModule) {
	r.Modules = make([]ResponseProducerModule, len(list))
	for i, e := range list {
		r.Modules[i].FromEntity(e)
	}
}
