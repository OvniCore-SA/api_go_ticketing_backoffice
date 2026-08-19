package handlers

import (
	"strings"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/public_storefront"
	"github.com/gofiber/fiber/v2"
)

// PublicStorefrontHandler expone la vista pública (sin auth) que consume
// el frontend Next.js con el Vercel Platforms Starter Kit.
//
// Todas las respuestas incluyen Cache-Control para permitir caché en Edge
// (Vercel) y aliviar carga sobre el backend. Los TTL son cortos porque el
// SuperAdmin cambia la config con relativa frecuencia — 60s es el
// compromiso entre frescura y latencia.
type PublicStorefrontHandler struct {
	service public_storefront.Service
}

func NewPublicStorefrontHandler(service public_storefront.Service) *PublicStorefrontHandler {
	return &PublicStorefrontHandler{service: service}
}

const (
	// cacheShort es para /resolve — el mapping dominio→producer cambia
	// pocas veces (solo cuando se registra o borra un dominio).
	cacheShort = "public, max-age=300, s-maxage=300"
	// cacheHot es para /storefront — el SuperAdmin puede publicar cambios
	// de la home con Puck; queremos que se vean en <1min.
	cacheHot = "public, max-age=60, s-maxage=60"
)

// ResolveDomain — GET /public/resolve/:domain
// Devuelve el mapping dominio → producer para que el middleware Edge de
// Next.js pueda cachearlo en Redis.
func (h *PublicStorefrontHandler) ResolveDomain(c *fiber.Ctx) error {
	domain := strings.ToLower(strings.TrimSpace(c.Params("domain")))
	result, err := h.service.ResolveDomainService(c.Context(), domain)
	if err != nil {
		return handleServiceError(c, err)
	}
	c.Set(fiber.HeaderCacheControl, cacheShort)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

// GetConfigBySlug — GET /public/storefront/:slug
// Devuelve el EffectiveConfig completo del producer (sanitizado).
func (h *PublicStorefrontHandler) GetConfigBySlug(c *fiber.Ctx) error {
	slug := strings.ToLower(strings.TrimSpace(c.Params("slug")))
	result, err := h.service.GetConfigBySlugService(c.Context(), slug)
	if err != nil {
		return handleServiceError(c, err)
	}
	c.Set(fiber.HeaderCacheControl, cacheHot)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

// GetPageBySlug — GET /public/storefront/:slug/pages/:pageType
// Devuelve solo la página solicitada. Es el endpoint más caliente —
// se llama en cada render server-side del frontend.
func (h *PublicStorefrontHandler) GetPageBySlug(c *fiber.Ctx) error {
	slug := strings.ToLower(strings.TrimSpace(c.Params("slug")))
	pageType := strings.ToLower(strings.TrimSpace(c.Params("pageType")))
	result, err := h.service.GetPageBySlugService(c.Context(), slug, pageType)
	if err != nil {
		return handleServiceError(c, err)
	}
	c.Set(fiber.HeaderCacheControl, cacheHot)
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}
