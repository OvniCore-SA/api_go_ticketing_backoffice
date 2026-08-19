package moduledtos

import "github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"

type ResponseModule struct {
	ID          uint   `json:"id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
	IsCore      bool   `json:"is_core"`
}

func (r *ResponseModule) FromEntity(e entities.Module) {
	r.ID = e.ID
	r.Code = e.Code
	r.Name = e.Name
	r.Description = e.Description
	r.Category = e.Category
	r.IsCore = e.IsCore
}

type ResponseModules struct {
	Modules []ResponseModule `json:"modules"`
}

func (r *ResponseModules) FromEntities(list []entities.Module) {
	r.Modules = make([]ResponseModule, len(list))
	for i, e := range list {
		r.Modules[i].FromEntity(e)
	}
}

type ResponseComponentVariant struct {
	ID          uint   `json:"id"`
	ModuleID    uint   `json:"module_id"`
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PreviewURL  string `json:"preview_url"`
}

func (r *ResponseComponentVariant) FromEntity(e entities.ComponentVariant) {
	r.ID = e.ID
	r.ModuleID = e.ModuleID
	r.Code = e.Code
	r.Name = e.Name
	r.Description = e.Description
	r.PreviewURL = e.PreviewURL
}

type ResponseComponentVariants struct {
	Variants []ResponseComponentVariant `json:"variants"`
}

func (r *ResponseComponentVariants) FromEntities(list []entities.ComponentVariant) {
	r.Variants = make([]ResponseComponentVariant, len(list))
	for i, e := range list {
		r.Variants[i].FromEntity(e)
	}
}
