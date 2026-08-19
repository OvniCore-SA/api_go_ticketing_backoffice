package handlers

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/auth"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/authdtos"
	"github.com/gofiber/fiber/v2"
)

// authUserFromCtx extrae el SuperAdmin autenticado seteado por ValidateToken.
func authUserFromCtx(c *fiber.Ctx) authdtos.ResponseUser {
	user, _ := c.Locals("user").(authdtos.ResponseUser)
	return user
}

// handleAuthError traduce errores del dominio auth a respuestas HTTP.
func handleAuthError(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": false, "message": err.Error()})
	case errors.Is(err, auth.ErrAccountInactive):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": false, "message": err.Error()})
	case errors.Is(err, auth.ErrNotSuperAdmin):
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": false, "message": err.Error()})
	}

	if authErr, ok := err.(*auth.AuthError); ok {
		if authErr.Code == fiber.StatusInternalServerError {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": false, "message": auth.MsgInternalError})
		}
		return c.Status(authErr.Code).JSON(fiber.Map{"status": false, "message": authErr.Message})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": false, "message": auth.MsgInternalError})
}

// handleServiceError es el handler genérico para dominios que devuelven
// fiber.Error (patrón MI_ESTILO_GO_BACKEND.md).
func handleServiceError(c *fiber.Ctx, err error) error {
	if fiberErr, ok := err.(*fiber.Error); ok {
		return c.Status(fiberErr.Code).JSON(fiber.Map{
			"status":  false,
			"message": fiberErr.Message,
		})
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"status":  false,
		"message": err.Error(),
	})
}
