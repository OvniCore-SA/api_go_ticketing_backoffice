package storedtos

import (
	"encoding/json"
	"errors"
)

var (
	ErrHostRequired   = errors.New("el parámetro host o dominio es requerido")
	ErrTenantNotFound = errors.New("tenant o dominio no encontrado")
)

type TenantMeta struct {
	ID           uint   `json:"id"`
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	ContactEmail string `json:"contact_email,omitempty"`
	Status       string `json:"status"`
	Suspended    bool   `json:"suspended"`
	TemplateID   *uint  `json:"template_id,omitempty"`
	PlanID       *uint  `json:"plan_id,omitempty"`
}

type DomainMeta struct {
	Domain   string `json:"domain"`
	Type     string `json:"type"`
	Verified bool   `json:"verified"`
}

type ActiveModule struct {
	ID     uint   `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	IsCore bool   `json:"is_core"`
}

type ActiveVariant struct {
	ModuleID           uint   `json:"module_id"`
	ComponentVariantID uint   `json:"component_variant_id"`
	VariantCode        string `json:"variant_code,omitempty"`
}

type ActiveTokens struct {
	Colors  json.RawMessage `json:"colors,omitempty"`
	Fonts   json.RawMessage `json:"fonts,omitempty"`
	Radius  json.RawMessage `json:"radius,omitempty"`
	Shadows json.RawMessage `json:"shadows,omitempty"`
}

type ResponseTenantByHost struct {
	Tenant        TenantMeta                 `json:"tenant"`
	Domain        DomainMeta                 `json:"domain"`
	ActiveModules []ActiveModule             `json:"active_modules"`
	Modules       []ActiveModule             `json:"modules"`
	Variants      map[uint]ActiveVariant     `json:"variants,omitempty"`
	Tokens        ActiveTokens               `json:"tokens,omitempty"`
	Pages         map[string]json.RawMessage `json:"pages,omitempty"`
}
