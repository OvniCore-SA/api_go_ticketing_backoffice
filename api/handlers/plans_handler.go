package handlers

import (
	"strconv"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/plans"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/plandtos"
	"github.com/gofiber/fiber/v2"
)

type PlansHandler struct {
	service plans.Service
}

func NewPlansHandler(service plans.Service) *PlansHandler {
	return &PlansHandler{service: service}
}

func (h *PlansHandler) Create(c *fiber.Ctx) error {
	var req plandtos.RequestCreatePlan
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
	}
	result, err := h.service.CreatePlanService(req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": true, "data": result})
}

func (h *PlansHandler) List(c *fiber.Ctx) error {
	var req plandtos.RequestListPlans
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "query params inválidos"})
	}
	list, meta, err := h.service.GetPlansService(req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": list, "meta": meta})
}

func (h *PlansHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "ID inválido"})
	}
	result, err := h.service.GetPlanByIDService(uint(id))
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *PlansHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "ID inválido"})
	}
	var req plandtos.RequestUpdatePlan
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
	}
	result, err := h.service.UpdatePlanService(uint(id), req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *PlansHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "ID inválido"})
	}
	if err := h.service.DeletePlanService(uint(id)); err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "message": "plan eliminado"})
}
