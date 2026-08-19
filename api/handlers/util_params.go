package handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// parseUintParam parsea un param de path como uint>0 y responde 400 si falla.
// Se aísla acá para no repetir el mismo bloque en cada handler.
func parseUintParam(c *fiber.Ctx, name string) (uint, error) {
	v, err := strconv.ParseUint(c.Params(name), 10, 64)
	if err != nil || v == 0 {
		return 0, c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"status":  false,
			"message": name + " inválido",
		})
	}
	return uint(v), nil
}
