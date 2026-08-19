package routes

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api/handlers"
	"github.com/gofiber/fiber/v2"
)

func SetupModulesRoutes(router fiber.Router, handler *handlers.ModulesHandler, authMiddleware fiber.Handler) {
	g := router.Group("/modules", authMiddleware)
	g.Get("/", handler.List)
	g.Get("/:id/variants", handler.GetVariants)
}
