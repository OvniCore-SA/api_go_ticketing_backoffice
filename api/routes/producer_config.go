package routes

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api/handlers"
	"github.com/gofiber/fiber/v2"
)

// SetupProducerConfigRoutes monta el endpoint de configuración efectiva.
// Va bajo /producers/:id/... como los otros sub-recursos, pero se separa
// en su propia función porque no maneja escritura — solo lectura resuelta.
func SetupProducerConfigRoutes(router fiber.Router, handler *handlers.ProducerConfigHandler, authMiddleware fiber.Handler) {
	g := router.Group("/producers", authMiddleware)
	g.Get("/:id/effective-config", handler.GetEffectiveConfig)
}
