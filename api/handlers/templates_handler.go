package handlers

import (
	"strconv"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/templates"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/templatedtos"
	"github.com/gofiber/fiber/v2"
)

type TemplatesHandler struct {
	service templates.Service
}

func NewTemplatesHandler(service templates.Service) *TemplatesHandler {
	return &TemplatesHandler{service: service}
}

func (h *TemplatesHandler) Create(c *fiber.Ctx) error {
	var req templatedtos.RequestCreateTemplate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
	}
	result, err := h.service.CreateTemplateService(req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": true, "data": result})
}

func (h *TemplatesHandler) List(c *fiber.Ctx) error {
	var req templatedtos.RequestListTemplates
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "query params inválidos"})
	}
	list, meta, err := h.service.GetTemplatesService(req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": list, "meta": meta})
}

func (h *TemplatesHandler) GetByID(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "ID inválido"})
	}
	result, err := h.service.GetTemplateByIDService(uint(id))
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *TemplatesHandler) Update(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "ID inválido"})
	}
	var req templatedtos.RequestUpdateTemplate
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
	}
	result, err := h.service.UpdateTemplateService(uint(id), req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *TemplatesHandler) Delete(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "ID inválido"})
	}
	if err := h.service.DeleteTemplateService(uint(id)); err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "message": "template eliminada"})
}

func (h *TemplatesHandler) GetModules(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "ID inválido"})
	}
	result, err := h.service.GetTemplateModulesService(uint(id))
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *TemplatesHandler) ReplaceModules(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "ID inválido"})
	}
	var req templatedtos.RequestReplaceTemplateModules
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
	}
	result, err := h.service.ReplaceTemplateModulesService(uint(id), req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *TemplatesHandler) GetPages(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "ID inválido"})
	}
	result, err := h.service.GetTemplatePagesService(uint(id))
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *TemplatesHandler) UpsertPage(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "ID inválido"})
	}
	pageType := c.Params("pageType")
	var req templatedtos.RequestUpsertTemplatePage
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
	}
	result, err := h.service.UpsertTemplatePageService(uint(id), pageType, req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}
