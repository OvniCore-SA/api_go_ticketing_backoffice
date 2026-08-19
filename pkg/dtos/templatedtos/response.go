package templatedtos

import (
	"encoding/json"
	"time"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
)

type ResponseTemplate struct {
	ID             uint            `json:"id"`
	Code           string          `json:"code"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	PreviewURL     string          `json:"preview_url"`
	DefaultColors  json.RawMessage `json:"default_colors,omitempty"`
	DefaultFonts   json.RawMessage `json:"default_fonts,omitempty"`
	DefaultRadius  json.RawMessage `json:"default_radius,omitempty"`
	DefaultShadows json.RawMessage `json:"default_shadows,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func (r *ResponseTemplate) FromEntity(e entities.Template) {
	r.ID = e.ID
	r.Code = e.Code
	r.Name = e.Name
	r.Description = e.Description
	r.PreviewURL = e.PreviewURL
	r.DefaultColors = e.DefaultColors
	r.DefaultFonts = e.DefaultFonts
	r.DefaultRadius = e.DefaultRadius
	r.DefaultShadows = e.DefaultShadows
	r.CreatedAt = e.CreatedAt
	r.UpdatedAt = e.UpdatedAt
}

type ResponseTemplates struct {
	Templates []ResponseTemplate `json:"templates"`
}

func (r *ResponseTemplates) FromEntities(list []entities.Template) {
	r.Templates = make([]ResponseTemplate, len(list))
	for i, e := range list {
		r.Templates[i].FromEntity(e)
	}
}

type ResponseTemplateModule struct {
	ID         uint `json:"id"`
	TemplateID uint `json:"template_id"`
	ModuleID   uint `json:"module_id"`
	ModuleCode string `json:"module_code,omitempty"`
}

func (r *ResponseTemplateModule) FromEntity(e entities.TemplateModule) {
	r.ID = e.ID
	r.TemplateID = e.TemplateID
	r.ModuleID = e.ModuleID
	r.ModuleCode = e.Module.Code
}

type ResponseTemplateModules struct {
	Items []ResponseTemplateModule `json:"items"`
}

func (r *ResponseTemplateModules) FromEntities(list []entities.TemplateModule) {
	r.Items = make([]ResponseTemplateModule, len(list))
	for i, e := range list {
		r.Items[i].FromEntity(e)
	}
}

type ResponseTemplatePage struct {
	ID              uint            `json:"id"`
	TemplateID      uint            `json:"template_id"`
	PageType        string          `json:"page_type"`
	PuckJSONDefault json.RawMessage `json:"puck_json_default,omitempty"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

func (r *ResponseTemplatePage) FromEntity(e entities.TemplatePage) {
	r.ID = e.ID
	r.TemplateID = e.TemplateID
	r.PageType = e.PageType
	r.PuckJSONDefault = e.PuckJSONDefault
	r.UpdatedAt = e.UpdatedAt
}
