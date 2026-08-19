package handlers

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/commissions"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/design_tokens"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/domains"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/page_templates"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producer_component_variants"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producer_modules"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producers"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/commissiondtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/designtokendtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/domaindtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/pagetemplatedtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/producerdtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/producermoduledtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/producervariantdtos"
	"github.com/gofiber/fiber/v2"
)

// ProducersHandler concentra los endpoints del producer y de sus sub-recursos
// bajo el mismo prefijo /producers/:id/... — todos los sub-servicios ya
// validan la existencia del producer antes de operar.
type ProducersHandler struct {
	producersSvc producers.Service
	modulesSvc   producer_modules.Service
	variantsSvc  producer_component_variants.Service
	tokensSvc    design_tokens.Service
	pagesSvc     page_templates.Service
	domainsSvc   domains.Service
	commissSvc   commissions.Service
}

func NewProducersHandler(
	producersSvc producers.Service,
	modulesSvc producer_modules.Service,
	variantsSvc producer_component_variants.Service,
	tokensSvc design_tokens.Service,
	pagesSvc page_templates.Service,
	domainsSvc domains.Service,
	commissSvc commissions.Service,
) *ProducersHandler {
	return &ProducersHandler{
		producersSvc: producersSvc,
		modulesSvc:   modulesSvc,
		variantsSvc:  variantsSvc,
		tokensSvc:    tokensSvc,
		pagesSvc:     pagesSvc,
		domainsSvc:   domainsSvc,
		commissSvc:   commissSvc,
	}
}

// ─── Producer CRUD ────────────────────────────────────────────────────────────

func (h *ProducersHandler) Create(c *fiber.Ctx) error {
	var req producerdtos.RequestCreateProducer
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
	}
	result, err := h.producersSvc.CreateProducerService(req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": true, "data": result})
}

func (h *ProducersHandler) List(c *fiber.Ctx) error {
	var req producerdtos.RequestListProducers
	if err := c.QueryParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "query params inválidos"})
	}
	list, meta, err := h.producersSvc.GetProducersService(req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": list, "meta": meta})
}

func (h *ProducersHandler) GetByID(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	result, err := h.producersSvc.GetProducerByIDService(id)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *ProducersHandler) Update(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	var req producerdtos.RequestUpdateProducer
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
	}
	result, err := h.producersSvc.UpdateProducerService(id, req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *ProducersHandler) Delete(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	if err := h.producersSvc.DeleteProducerService(id); err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "message": "producer eliminado"})
}

// ─── Producer modules ─────────────────────────────────────────────────────────

func (h *ProducersHandler) ListModules(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	result, err := h.modulesSvc.ListByProducerService(id)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *ProducersHandler) ToggleModule(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	moduleID, err := parseUintParam(c, "moduleId")
	if err != nil {
		return err
	}
	var req producermoduledtos.RequestToggleModule
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
	}
	result, err := h.modulesSvc.ToggleModuleService(id, moduleID, req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

// ─── Component variants por producer ──────────────────────────────────────────

func (h *ProducersHandler) ListVariants(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	result, err := h.variantsSvc.ListByProducerService(id)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *ProducersHandler) AssignVariant(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	moduleID, err := parseUintParam(c, "moduleId")
	if err != nil {
		return err
	}
	var req producervariantdtos.RequestAssignVariant
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
	}
	result, err := h.variantsSvc.AssignVariantService(id, moduleID, req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

// ─── Design tokens ────────────────────────────────────────────────────────────

func (h *ProducersHandler) GetTokens(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	result, err := h.tokensSvc.GetTokensService(id)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *ProducersHandler) UpdateTokens(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	var req designtokendtos.RequestUpdateDesignTokens
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
	}
	result, err := h.tokensSvc.UpdateTokensService(id, req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

// ─── Page templates ───────────────────────────────────────────────────────────

func (h *ProducersHandler) ListPages(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	result, err := h.pagesSvc.ListByProducerService(id)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *ProducersHandler) GetPage(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	result, err := h.pagesSvc.GetOneService(id, c.Params("pageType"))
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *ProducersHandler) SavePage(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	var req pagetemplatedtos.RequestSavePage
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
	}
	result, err := h.pagesSvc.SavePageService(id, c.Params("pageType"), req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

// ─── Domains ──────────────────────────────────────────────────────────────────

func (h *ProducersHandler) ListDomains(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	result, err := h.domainsSvc.ListByProducerService(id)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *ProducersHandler) CreateDomain(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	var req domaindtos.RequestCreateDomain
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
	}
	result, err := h.domainsSvc.CreateService(id, req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": true, "data": result})
}

func (h *ProducersHandler) DeleteDomain(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	domainID, err := parseUintParam(c, "domainId")
	if err != nil {
		return err
	}
	if err := h.domainsSvc.DeleteService(id, domainID); err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "message": "dominio eliminado"})
}

// ─── Commissions ──────────────────────────────────────────────────────────────

func (h *ProducersHandler) ListCommissions(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	result, err := h.commissSvc.ListByProducerService(id)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}

func (h *ProducersHandler) CreateCommission(c *fiber.Ctx) error {
	id, err := parseUintParam(c, "id")
	if err != nil {
		return err
	}
	var req commissiondtos.RequestCreateCommission
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
	}
	result, err := h.commissSvc.CreateService(id, req)
	if err != nil {
		return handleServiceError(c, err)
	}
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": true, "data": result})
}
