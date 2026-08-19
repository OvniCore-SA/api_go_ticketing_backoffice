package routes

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api/handlers"
	"github.com/gofiber/fiber/v2"
)

// SetupAuthRoutes monta los endpoints de autenticación del backoffice.
// El único endpoint público es /auth/login; el resto exige token válido.
func SetupAuthRoutes(router fiber.Router, handler *handlers.AuthHandler, authMiddleware fiber.Handler) {
	group := router.Group("/auth")

	group.Post("/login", handler.Login)

	protected := group.Group("", authMiddleware)
	protected.Get("/me", handler.Me)
	protected.Post("/refresh", handler.Refresh)
	protected.Put("/change-password", handler.ChangePassword)
	protected.Post("/logout", handler.Logout)
}
