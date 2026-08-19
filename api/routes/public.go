package routes

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api/handlers"
	"github.com/gofiber/fiber/v2"
)

// SetupPublicStorefrontRoutes monta los endpoints sin autenticación que
// consume el frontend Next.js. Van bajo /public para separarlos
// claramente del resto de la API (que requiere JWT SuperAdmin).
//
// El caller de esta función es el responsable de aplicar CORS abierto
// y rate limiter al grupo — se hace en app.go, no acá.
func SetupPublicStorefrontRoutes(router fiber.Router, handler *handlers.PublicStorefrontHandler) {
	g := router.Group("/public")

	g.Get("/resolve/:domain", handler.ResolveDomain)
	g.Get("/storefront/:slug", handler.GetConfigBySlug)
	g.Get("/storefront/:slug/pages/:pageType", handler.GetPageBySlug)
}
