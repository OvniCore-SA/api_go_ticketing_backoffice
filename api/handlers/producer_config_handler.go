package handlers

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producer_config"
	"github.com/gofiber/fiber/v2"
)

// ProducerConfigHandler expone la configuración efectiva resuelta del
// producer (resultado del pipeline Decorator). Sirve dos propósitos:
//
//  1. Panel del SuperAdmin: previsualizar exactamente qué va a ver el fan
//     antes de publicar cambios ("¿está bien esta home?").
//  2. Consumo por ticketing-core: mientras compartan DB, ticketing-core
//     puede leer las tablas directo (spec §5.2), pero cuando se separen
//     este endpoint queda como contrato estable.
type ProducerConfigHandler struct {
	resolver *producer_config.Resolver
}

func NewProducerConfigHandler(resolver *producer_config.Resolver) *ProducerConfigHandler {
	return &ProducerConfigHandler{resolver: resolver}
}

// GetEffectiveConfig — GET /producers/:id/effective-config
func (h *ProducerConfigHandler) GetEffectiveConfig(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	cfg, err := h.resolver.Resolve(c.Context(), id)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": cfg})
}
