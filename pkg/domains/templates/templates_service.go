package templates

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/templatedtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/utils"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Service interface {
	CreateTemplateService(req templatedtos.RequestCreateTemplate) (templatedtos.ResponseTemplate, error)
	GetTemplatesService(req templatedtos.RequestListTemplates) (templatedtos.ResponseTemplates, utils.Pagination, error)
	GetTemplateByIDService(id uint) (templatedtos.ResponseTemplate, error)
	UpdateTemplateService(id uint, req templatedtos.RequestUpdateTemplate) (templatedtos.ResponseTemplate, error)
	DeleteTemplateService(id uint) error

	GetTemplateModulesService(templateID uint) (templatedtos.ResponseTemplateModules, error)
	ReplaceTemplateModulesService(templateID uint, req templatedtos.RequestReplaceTemplateModules) (templatedtos.ResponseTemplateModules, error)

	GetTemplatePagesService(templateID uint) ([]templatedtos.ResponseTemplatePage, error)
	UpsertTemplatePageService(templateID uint, pageType string, req templatedtos.RequestUpsertTemplatePage) (templatedtos.ResponseTemplatePage, error)
}

type service struct {
	repo Repository
}

func NewTemplatesService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreateTemplateService(req templatedtos.RequestCreateTemplate) (templatedtos.ResponseTemplate, error) {
	if err := req.Validate(); err != nil {
		return templatedtos.ResponseTemplate{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if _, err := s.repo.GetTemplateRepository(filters.TemplateFilter{Code: req.Code}); err == nil {
		return templatedtos.ResponseTemplate{}, fiber.NewError(fiber.StatusConflict, templatedtos.ErrTemplateCodeExists.Error())
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return templatedtos.ResponseTemplate{}, fiber.NewError(fiber.StatusInternalServerError, "error al validar código de la template")
	}

	created, err := s.repo.CreateTemplateRepository(req.ToEntity())
	if err != nil {
		return templatedtos.ResponseTemplate{}, fiber.NewError(fiber.StatusInternalServerError, "error al crear la template")
	}
	var response templatedtos.ResponseTemplate
	response.FromEntity(created)
	return response, nil
}

func (s *service) GetTemplatesService(req templatedtos.RequestListTemplates) (response templatedtos.ResponseTemplates, pagination utils.Pagination, err error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	list, total, err := s.repo.GetAllTemplatesRepository(filters.TemplateFilter{Search: req.Search}, req.Page, req.Limit)
	if err != nil {
		err = fiber.NewError(fiber.StatusInternalServerError, "error al obtener templates")
		return
	}
	response.FromEntities(list)
	pagination = utils.NewPagination(req.Page, req.Limit, total)
	return
}

func (s *service) GetTemplateByIDService(id uint) (templatedtos.ResponseTemplate, error) {
	entity, err := s.repo.GetTemplateRepository(filters.TemplateFilter{ID: id})
	if err != nil {
		return templatedtos.ResponseTemplate{}, fiber.NewError(fiber.StatusNotFound, templatedtos.ErrTemplateNotFound.Error())
	}
	var response templatedtos.ResponseTemplate
	response.FromEntity(entity)
	return response, nil
}

func (s *service) UpdateTemplateService(id uint, req templatedtos.RequestUpdateTemplate) (templatedtos.ResponseTemplate, error) {
	if err := req.Validate(); err != nil {
		return templatedtos.ResponseTemplate{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if _, err := s.repo.GetTemplateRepository(filters.TemplateFilter{ID: id}); err != nil {
		return templatedtos.ResponseTemplate{}, fiber.NewError(fiber.StatusNotFound, templatedtos.ErrTemplateNotFound.Error())
	}
	fields := map[string]interface{}{}
	if req.Name != "" {
		fields["name"] = req.Name
	}
	if req.Description != "" {
		fields["description"] = req.Description
	}
	if req.PreviewURL != "" {
		fields["preview_url"] = req.PreviewURL
	}
	if len(req.DefaultColors) > 0 {
		fields["default_colors"] = req.DefaultColors
	}
	if len(req.DefaultFonts) > 0 {
		fields["default_fonts"] = req.DefaultFonts
	}
	if len(req.DefaultRadius) > 0 {
		fields["default_radius"] = req.DefaultRadius
	}
	if len(req.DefaultShadows) > 0 {
		fields["default_shadows"] = req.DefaultShadows
	}
	if len(fields) > 0 {
		if err := s.repo.UpdateTemplateRepository(id, fields); err != nil {
			return templatedtos.ResponseTemplate{}, fiber.NewError(fiber.StatusInternalServerError, "error al actualizar la template")
		}
	}
	updated, _ := s.repo.GetTemplateRepository(filters.TemplateFilter{ID: id})
	var response templatedtos.ResponseTemplate
	response.FromEntity(updated)
	return response, nil
}

func (s *service) DeleteTemplateService(id uint) error {
	if _, err := s.repo.GetTemplateRepository(filters.TemplateFilter{ID: id}); err != nil {
		return fiber.NewError(fiber.StatusNotFound, templatedtos.ErrTemplateNotFound.Error())
	}

	// Guardarraíl blast-radius global: borrar una template referenciada
	// deja Producers con template_id colgado y sin capa de defaults en
	// el Resolver.
	count, err := s.repo.CountProducersByTemplateRepository(id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error al validar uso de la template")
	}
	if count > 0 {
		return fiber.NewError(fiber.StatusConflict, templatedtos.ErrTemplateInUse.Error())
	}

	if err := s.repo.DeleteTemplateRepository(id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error al eliminar la template")
	}
	return nil
}

func (s *service) GetTemplateModulesService(templateID uint) (templatedtos.ResponseTemplateModules, error) {
	if _, err := s.repo.GetTemplateRepository(filters.TemplateFilter{ID: templateID}); err != nil {
		return templatedtos.ResponseTemplateModules{}, fiber.NewError(fiber.StatusNotFound, templatedtos.ErrTemplateNotFound.Error())
	}
	list, err := s.repo.GetTemplateModulesRepository(templateID)
	if err != nil {
		return templatedtos.ResponseTemplateModules{}, fiber.NewError(fiber.StatusInternalServerError, "error al obtener módulos de la template")
	}
	var response templatedtos.ResponseTemplateModules
	response.FromEntities(list)
	return response, nil
}

func (s *service) ReplaceTemplateModulesService(templateID uint, req templatedtos.RequestReplaceTemplateModules) (templatedtos.ResponseTemplateModules, error) {
	if _, err := s.repo.GetTemplateRepository(filters.TemplateFilter{ID: templateID}); err != nil {
		return templatedtos.ResponseTemplateModules{}, fiber.NewError(fiber.StatusNotFound, templatedtos.ErrTemplateNotFound.Error())
	}
	if err := s.repo.ReplaceTemplateModulesRepository(templateID, req.ModuleIDs); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return templatedtos.ResponseTemplateModules{}, fiber.NewError(fiber.StatusBadRequest, templatedtos.ErrModuleNotFound.Error())
		}
		return templatedtos.ResponseTemplateModules{}, fiber.NewError(fiber.StatusInternalServerError, "error al actualizar módulos de la template")
	}
	list, _ := s.repo.GetTemplateModulesRepository(templateID)
	var response templatedtos.ResponseTemplateModules
	response.FromEntities(list)
	return response, nil
}

func (s *service) GetTemplatePagesService(templateID uint) ([]templatedtos.ResponseTemplatePage, error) {
	if _, err := s.repo.GetTemplateRepository(filters.TemplateFilter{ID: templateID}); err != nil {
		return nil, fiber.NewError(fiber.StatusNotFound, templatedtos.ErrTemplateNotFound.Error())
	}
	list, err := s.repo.GetTemplatePagesRepository(templateID)
	if err != nil {
		return nil, fiber.NewError(fiber.StatusInternalServerError, "error al obtener páginas de la template")
	}
	out := make([]templatedtos.ResponseTemplatePage, len(list))
	for i, e := range list {
		out[i].FromEntity(e)
	}
	return out, nil
}

func (s *service) UpsertTemplatePageService(templateID uint, pageType string, req templatedtos.RequestUpsertTemplatePage) (templatedtos.ResponseTemplatePage, error) {
	if err := req.Validate(); err != nil {
		return templatedtos.ResponseTemplatePage{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if !isValidPageType(pageType) {
		return templatedtos.ResponseTemplatePage{}, fiber.NewError(fiber.StatusBadRequest, templatedtos.ErrPageTypeInvalid.Error())
	}
	if _, err := s.repo.GetTemplateRepository(filters.TemplateFilter{ID: templateID}); err != nil {
		return templatedtos.ResponseTemplatePage{}, fiber.NewError(fiber.StatusNotFound, templatedtos.ErrTemplateNotFound.Error())
	}

	saved, err := s.repo.UpsertTemplatePageRepository(templateID, pageType, req.PuckJSON)
	if err != nil {
		return templatedtos.ResponseTemplatePage{}, fiber.NewError(fiber.StatusInternalServerError, "error al guardar la página de la template")
	}
	var response templatedtos.ResponseTemplatePage
	response.FromEntity(saved)
	return response, nil
}

func isValidPageType(pt string) bool {
	switch pt {
	case entities.PageTypeHome,
		entities.PageTypeEventDetail,
		entities.PageTypeCheckout,
		entities.PageTypeGallery,
		entities.PageTypeFAQ,
		entities.PageTypeContact:
		return true
	}
	return false
}
