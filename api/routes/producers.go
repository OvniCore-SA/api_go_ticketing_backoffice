package routes

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api/handlers"
	"github.com/gofiber/fiber/v2"
)

// SetupProducersRoutes monta el CRUD de producers y todos sus sub-recursos
// (módulos, variantes, tokens, páginas Puck, dominios, comisiones) bajo el
// mismo prefijo /producers/:id/... y con el mismo middleware de autenticación.
func SetupProducersRoutes(router fiber.Router, handler *handlers.ProducersHandler, authMiddleware fiber.Handler) {
	g := router.Group("/producers", authMiddleware)

	// CRUD del producer.
	g.Get("/", handler.List)
	g.Post("/", handler.Create)
	g.Get("/:id", handler.GetByID)
	g.Patch("/:id", handler.Update)
	g.Delete("/:id", handler.Delete)

	// Módulos habilitados.
	g.Get("/:id/modules", handler.ListModules)
	g.Put("/:id/modules/:moduleId", handler.ToggleModule)

	// Variantes elegidas por módulo.
	g.Get("/:id/component-variants", handler.ListVariants)
	g.Put("/:id/component-variants/:moduleId", handler.AssignVariant)

	// Design tokens.
	g.Get("/:id/design-tokens", handler.GetTokens)
	g.Put("/:id/design-tokens", handler.UpdateTokens)

	// Páginas Puck (onPublish del editor).
	g.Get("/:id/page-templates", handler.ListPages)
	g.Get("/:id/page-templates/:pageType", handler.GetPage)
	g.Put("/:id/page-templates/:pageType", handler.SavePage)

	// Dominios.
	g.Get("/:id/domains", handler.ListDomains)
	g.Post("/:id/domains", handler.CreateDomain)
	g.Delete("/:id/domains/:domainId", handler.DeleteDomain)

	// Comisiones.
	g.Get("/:id/commissions", handler.ListCommissions)
	g.Post("/:id/commissions", handler.CreateCommission)
}
