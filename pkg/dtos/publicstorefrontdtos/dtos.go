// Package publicstorefrontdtos define los payloads públicos que consume
// el frontend (Next.js + Vercel Platforms Starter Kit) sin autenticación.
//
// Estos DTOs son la vista slim del EffectiveConfig interno: eliminan
// metadatos de implementación (AppliedLayers, Source) y filtran módulos
// deshabilitados. Son un contrato estable — cuando este endpoint migre
// a ticketing-core, el shape se conserva.
package publicstorefrontdtos

import (
	"encoding/json"
	"errors"
)

var (
	ErrDomainNotFound   = errors.New("dominio no registrado")
	ErrProducerNotFound = errors.New("producer no encontrado")
	ErrPageNotFound     = errors.New("página no encontrada para este producer")
	ErrPageTypeInvalid  = errors.New("tipo de página no válido")
)

// ResponseResolveDomain es el payload que consume el middleware Edge de
// Next.js para mapear dominio → tenant. Se cachea en Redis con TTL 5min.
type ResponseResolveDomain struct {
	ProducerID uint   `json:"producer_id"`
	Slug       string `json:"slug"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Suspended  bool   `json:"suspended"`
	// DomainType permite al frontend saber si viene por subdomain o custom.
	DomainType string `json:"domain_type,omitempty"`
}

// ResponsePublicConfig es la vista pública del EffectiveConfig. Sin
// metadatos de debug ni sources — solo lo que el frontend necesita para
// renderizar.
type ResponsePublicConfig struct {
	Producer PublicProducerMeta         `json:"producer"`
	Modules  []PublicModule             `json:"modules"`  // solo enabled=true
	Variants map[uint]PublicVariant     `json:"variants"` // moduleID → variante
	Tokens   PublicTokens               `json:"tokens"`
	Pages    map[string]json.RawMessage `json:"pages"` // pageType → puck_json
}

// PublicProducerMeta expone solo lo necesario para branding + estado.
type PublicProducerMeta struct {
	ID        uint   `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Suspended bool   `json:"suspended"`
}

type PublicModule struct {
	ID     uint   `json:"id"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	IsCore bool   `json:"is_core"`
}

type PublicVariant struct {
	ModuleID           uint   `json:"module_id"`
	ComponentVariantID uint   `json:"component_variant_id"`
	VariantCode        string `json:"variant_code,omitempty"`
}

type PublicTokens struct {
	Colors  json.RawMessage `json:"colors,omitempty"`
	Fonts   json.RawMessage `json:"fonts,omitempty"`
	Radius  json.RawMessage `json:"radius,omitempty"`
	Shadows json.RawMessage `json:"shadows,omitempty"`
}

// ResponsePublicPage es la respuesta cuando el frontend pide una única
// página (típicamente en getServerSideProps por pageType).
type ResponsePublicPage struct {
	Producer PublicProducerMeta `json:"producer"`
	PageType string             `json:"page_type"`
	PuckJSON json.RawMessage    `json:"puck_json"`
	Tokens   PublicTokens       `json:"tokens"`
	Modules  []PublicModule     `json:"modules"`
}
