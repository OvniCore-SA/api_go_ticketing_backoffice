package routes

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api/handlers"
	"github.com/gofiber/fiber/v2"
)

func SetupPlansRoutes(router fiber.Router, handler *handlers.PlansHandler, authMiddleware fiber.Handler) {
	g := router.Group("/plans", authMiddleware)
	g.Get("/", handler.List)
	g.Post("/", handler.Create)
	g.Get("/:id", handler.GetByID)
	g.Patch("/:id", handler.Update)
	g.Delete("/:id", handler.Delete)
}
