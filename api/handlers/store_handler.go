package handlers

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/store"
	"github.com/gofiber/fiber/v2"
)

type StoreHandler struct {
	service store.Service
}

func NewStoreHandler(service store.Service) *StoreHandler {
	return &StoreHandler{service: service}
}

// GetTenantByHost — GET /api/v1/store/tenant-by-host
// Resuelve dominios/subdominios contra PostgreSQL y devuelve la configuración
// del tenant y sus módulos activos.
func (h *StoreHandler) GetTenantByHost(c *fiber.Ctx) error {
	host := c.Params("host")
	if host == "" {
		host = c.Query("host")
	}
	if host == "" {
		host = c.Query("domain")
	}
	if host == "" {
		host = c.Get("X-Forwarded-Host")
	}
	if host == "" {
		host = c.Get("Host")
	}

	result, err := h.service.GetTenantByHostService(c.Context(), host)
	if err != nil {
		return handleServiceError(c, err)
	}

	c.Set(fiber.HeaderCacheControl, "public, max-age=60, s-maxage=60")
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": true,
		"data":   result,
	})
}
