package middlewares

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

// NewLimiter crea un rate limiter por IP con respuesta JSON estándar.
func NewLimiter(max int, window time.Duration) fiber.Handler {
	return limiter.New(limiter.Config{
		Max:        max,
		Expiration: window,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
		LimitReached: func(c *fiber.Ctx) error {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"status":  false,
				"message": "demasiadas solicitudes, intentá más tarde",
			})
		},
	})
}

var (
	// StrictLimit — endpoints de alta sensibilidad: login, registro, contraseñas
	StrictLimit = NewLimiter(10, time.Minute)

	// EmailLimit — endpoints de envío de email: previene flooding
	EmailLimit = NewLimiter(5, time.Minute)

	// VerifyLimit — endpoints de verificación y consulta
	VerifyLimit = NewLimiter(20, time.Minute)

	// RefreshLimit — refresh de tokens
	RefreshLimit = NewLimiter(30, time.Minute)
)
