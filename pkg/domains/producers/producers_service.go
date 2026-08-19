package producers

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/plans"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/templates"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/producerdtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/utils"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Service interface {
	CreateProducerService(req producerdtos.RequestCreateProducer) (producerdtos.ResponseProducer, error)
	GetProducersService(req producerdtos.RequestListProducers) (producerdtos.ResponseProducers, utils.Pagination, error)
	GetProducerByIDService(id uint) (producerdtos.ResponseProducer, error)
	UpdateProducerService(id uint, req producerdtos.RequestUpdateProducer) (producerdtos.ResponseProducer, error)
	DeleteProducerService(id uint) error
}

type service struct {
	repo         Repository
	plansRepo    plans.Repository
	templatesSvc templates.Service // solo para validar existencia por ID
}

func NewProducersService(repo Repository, plansRepo plans.Repository, templatesSvc templates.Service) Service {
	return &service{
		repo:         repo,
		plansRepo:    plansRepo,
		templatesSvc: templatesSvc,
	}
}

func (s *service) CreateProducerService(req producerdtos.RequestCreateProducer) (producerdtos.ResponseProducer, error) {
	if err := req.Validate(); err != nil {
		return producerdtos.ResponseProducer{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if _, err := s.repo.GetProducerRepository(filters.ProducerFilter{Slug: req.Slug}); err == nil {
		return producerdtos.ResponseProducer{}, fiber.NewError(fiber.StatusConflict, producerdtos.ErrSlugExists.Error())
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return producerdtos.ResponseProducer{}, fiber.NewError(fiber.StatusInternalServerError, "error al validar slug")
	}

	if req.TemplateID != nil && *req.TemplateID > 0 {
		if _, err := s.templatesSvc.GetTemplateByIDService(*req.TemplateID); err != nil {
			return producerdtos.ResponseProducer{}, fiber.NewError(fiber.StatusBadRequest, producerdtos.ErrTemplateNotFound.Error())
		}
	}
	if req.PlanID != nil && *req.PlanID > 0 {
		if _, err := s.plansRepo.GetPlanRepository(filters.PlanFilter{ID: *req.PlanID}); err != nil {
			return producerdtos.ResponseProducer{}, fiber.NewError(fiber.StatusBadRequest, producerdtos.ErrPlanNotFound.Error())
		}
	}

	created, err := s.repo.SeedProducerFromTemplateRepository(req.ToEntity(), req.TemplateID)
	if err != nil {
		return producerdtos.ResponseProducer{}, fiber.NewError(fiber.StatusInternalServerError, "error al crear el producer")
	}

	var response producerdtos.ResponseProducer
	response.FromEntity(created)
	return response, nil
}

func (s *service) GetProducersService(req producerdtos.RequestListProducers) (response producerdtos.ResponseProducers, pagination utils.Pagination, err error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}
	list, total, err := s.repo.GetAllProducersRepository(filters.ProducerFilter{
		Search: req.Search,
		Status: req.Status,
		PlanID: req.PlanID,
	}, req.Page, req.Limit)
	if err != nil {
		err = fiber.NewError(fiber.StatusInternalServerError, "error al obtener producers")
		return
	}
	response.FromEntities(list)
	pagination = utils.NewPagination(req.Page, req.Limit, total)
	return
}

func (s *service) GetProducerByIDService(id uint) (producerdtos.ResponseProducer, error) {
	entity, err := s.repo.GetProducerRepository(filters.ProducerFilter{ID: id})
	if err != nil {
		return producerdtos.ResponseProducer{}, fiber.NewError(fiber.StatusNotFound, producerdtos.ErrProducerNotFound.Error())
	}
	var response producerdtos.ResponseProducer
	response.FromEntity(entity)
	return response, nil
}

func (s *service) UpdateProducerService(id uint, req producerdtos.RequestUpdateProducer) (producerdtos.ResponseProducer, error) {
	if err := req.Validate(); err != nil {
		return producerdtos.ResponseProducer{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if _, err := s.repo.GetProducerRepository(filters.ProducerFilter{ID: id}); err != nil {
		return producerdtos.ResponseProducer{}, fiber.NewError(fiber.StatusNotFound, producerdtos.ErrProducerNotFound.Error())
	}
	if req.PlanID != nil && *req.PlanID > 0 {
		if _, err := s.plansRepo.GetPlanRepository(filters.PlanFilter{ID: *req.PlanID}); err != nil {
			return producerdtos.ResponseProducer{}, fiber.NewError(fiber.StatusBadRequest, producerdtos.ErrPlanNotFound.Error())
		}
	}

	fields := map[string]interface{}{}
	if req.Name != "" {
		fields["name"] = req.Name
	}
	if req.ContactEmail != "" {
		fields["contact_email"] = req.ContactEmail
	}
	if req.Status != "" {
		fields["status"] = req.Status
	}
	if req.PlanID != nil {
		fields["plan_id"] = *req.PlanID
	}

	if len(fields) > 0 {
		if err := s.repo.UpdateProducerRepository(id, fields); err != nil {
			return producerdtos.ResponseProducer{}, fiber.NewError(fiber.StatusInternalServerError, "error al actualizar el producer")
		}
	}
	updated, _ := s.repo.GetProducerRepository(filters.ProducerFilter{ID: id})
	var response producerdtos.ResponseProducer
	response.FromEntity(updated)
	return response, nil
}

func (s *service) DeleteProducerService(id uint) error {
	if _, err := s.repo.GetProducerRepository(filters.ProducerFilter{ID: id}); err != nil {
		return fiber.NewError(fiber.StatusNotFound, producerdtos.ErrProducerNotFound.Error())
	}
	if err := s.repo.DeleteProducerRepository(id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error al eliminar el producer")
	}
	return nil
}
