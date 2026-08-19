package handlers

import (
	"strconv"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/modules"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/moduledtos"
	"github.com/gofiber/fiber/v2"
)

type ModulesHandler struct {
	service modules.Service
}

func NewModulesHandler(service modules.Service) *ModulesHandler {
	return &ModulesHandler{service: service}
}

func (h *ModulesHandler) List(c *fiber.Ctx) error {
	var req moduledtos.RequestListModules
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "query params inválidos"})
	}
	result, err := h.service.GetModulesService(req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *ModulesHandler) GetVariants(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "ID inválido"})
	}
	result, err := h.service.GetVariantsByModuleService(uint(id))
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}
