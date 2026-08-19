package routes

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api/handlers"
	"github.com/gofiber/fiber/v2"
)

func SetupTemplatesRoutes(router fiber.Router, handler *handlers.TemplatesHandler, authMiddleware fiber.Handler) {
	g := router.Group("/templates", authMiddleware)
	g.Get("/", handler.List)
	g.Post("/", handler.Create)
	g.Get("/:id", handler.GetByID)
	g.Patch("/:id", handler.Update)
	g.Delete("/:id", handler.Delete)

	g.Get("/:id/modules", handler.GetModules)
	g.Put("/:id/modules", handler.ReplaceModules)

	g.Get("/:id/pages", handler.GetPages)
	g.Put("/:id/pages/:pageType", handler.UpsertPage)
}
