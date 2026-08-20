package routes

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api/handlers"
	"github.com/gofiber/fiber/v2"
)

// SetupStoreRoutes monta los endpoints públicos del store engine bajo el prefijo /store.
func SetupStoreRoutes(router fiber.Router, handler *handlers.StoreHandler) {
	g := router.Group("/store")

	g.Get("/tenant-by-host", handler.GetTenantByHost)
	g.Get("/tenant-by-host/:host", handler.GetTenantByHost)
	g.Get("/tenant", handler.GetTenantByHost)
}
