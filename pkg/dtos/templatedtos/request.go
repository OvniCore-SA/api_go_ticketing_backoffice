package templatedtos

import (
	"encoding/json"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/commons"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
)

type RequestCreateTemplate struct {
	Code           string          `json:"code"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	PreviewURL     string          `json:"preview_url"`
	DefaultColors  json.RawMessage `json:"default_colors"`
	DefaultFonts   json.RawMessage `json:"default_fonts"`
	DefaultRadius  json.RawMessage `json:"default_radius"`
	DefaultShadows json.RawMessage `json:"default_shadows"`
}

func (r RequestCreateTemplate) Validate() error {
	if commons.StringIsEmpty(r.Code) {
		return ErrTemplateCodeReq
	}
	if commons.StringIsEmpty(r.Name) {
		return ErrTemplateNameReq
	}
	// Los cuatro jsonb defaults deben ser JSON válido si vienen.
	for _, raw := range []json.RawMessage{r.DefaultColors, r.DefaultFonts, r.DefaultRadius, r.DefaultShadows} {
		if len(raw) > 0 && !json.Valid(raw) {
			return ErrTemplateJSONBroken
		}
	}
	return nil
}

// Nota: el campo `code` es la clave lógica de la template — es INMUTABLE
// una vez creada. Por eso RequestUpdateTemplate no lo expone. Si necesitás
// renombrar una template, editás su `name`; el `code` queda como identidad
// estable para el resto del sistema (seed, referencias, etc.).

// Validate para el update: chequea que los jsonb sean válidos si vienen.
func (r RequestUpdateTemplate) Validate() error {
	for _, raw := range []json.RawMessage{r.DefaultColors, r.DefaultFonts, r.DefaultRadius, r.DefaultShadows} {
		if len(raw) > 0 && !json.Valid(raw) {
			return ErrTemplateJSONBroken
		}
	}
	return nil
}

// Validate para el upsert de página en una template.
func (r RequestUpsertTemplatePage) Validate() error {
	if len(r.PuckJSON) == 0 {
		return ErrTemplateJSONBroken
	}
	if !json.Valid(r.PuckJSON) {
		return ErrTemplateJSONBroken
	}
	return nil
}

func (r *RequestCreateTemplate) ToEntity() entities.Template {
	return entities.Template{
		Code:           r.Code,
		Name:           r.Name,
		Description:    r.Description,
		PreviewURL:     r.PreviewURL,
		DefaultColors:  r.DefaultColors,
		DefaultFonts:   r.DefaultFonts,
		DefaultRadius:  r.DefaultRadius,
		DefaultShadows: r.DefaultShadows,
	}
}

type RequestUpdateTemplate struct {
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	PreviewURL     string          `json:"preview_url"`
	DefaultColors  json.RawMessage `json:"default_colors"`
	DefaultFonts   json.RawMessage `json:"default_fonts"`
	DefaultRadius  json.RawMessage `json:"default_radius"`
	DefaultShadows json.RawMessage `json:"default_shadows"`
}

type RequestListTemplates struct {
	Search string `query:"search"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

type RequestReplaceTemplateModules struct {
	ModuleIDs []uint `json:"module_ids"`
}

type RequestUpsertTemplatePage struct {
	PuckJSON json.RawMessage `json:"puck_json"`
}
