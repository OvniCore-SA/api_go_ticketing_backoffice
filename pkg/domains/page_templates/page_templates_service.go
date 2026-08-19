package page_templates

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producers"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/pagetemplatedtos"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Service interface {
	ListByProducerService(producerID uint) (pagetemplatedtos.ResponsePageTemplates, error)
	GetOneService(producerID uint, pageType string) (pagetemplatedtos.ResponsePageTemplate, error)
	SavePageService(producerID uint, pageType string, req pagetemplatedtos.RequestSavePage) (pagetemplatedtos.ResponsePageTemplate, error)
}

type service struct {
	repo         Repository
	producersSvc producers.Service
}

func NewPageTemplatesService(repo Repository, producersSvc producers.Service) Service {
	return &service{repo: repo, producersSvc: producersSvc}
}

func (s *service) ListByProducerService(producerID uint) (pagetemplatedtos.ResponsePageTemplates, error) {
	if _, err := s.producersSvc.GetProducerByIDService(producerID); err != nil {
		return pagetemplatedtos.ResponsePageTemplates{}, err
	}
	list, err := s.repo.GetByProducerRepository(producerID)
	if err != nil {
		return pagetemplatedtos.ResponsePageTemplates{}, fiber.NewError(fiber.StatusInternalServerError, "error al obtener páginas")
	}
	var response pagetemplatedtos.ResponsePageTemplates
	response.FromEntities(list)
	return response, nil
}

func (s *service) GetOneService(producerID uint, pageType string) (pagetemplatedtos.ResponsePageTemplate, error) {
	if !pagetemplatedtos.IsValidPageType(pageType) {
		return pagetemplatedtos.ResponsePageTemplate{}, fiber.NewError(fiber.StatusBadRequest, pagetemplatedtos.ErrPageTypeInvalid.Error())
	}
	if _, err := s.producersSvc.GetProducerByIDService(producerID); err != nil {
		return pagetemplatedtos.ResponsePageTemplate{}, err
	}
	entity, err := s.repo.GetOneRepository(producerID, pageType)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pagetemplatedtos.ResponsePageTemplate{}, fiber.NewError(fiber.StatusNotFound, pagetemplatedtos.ErrPageNotFound.Error())
		}
		return pagetemplatedtos.ResponsePageTemplate{}, fiber.NewError(fiber.StatusInternalServerError, "error al obtener la página")
	}
	var response pagetemplatedtos.ResponsePageTemplate
	response.FromEntity(entity)
	return response, nil
}

func (s *service) SavePageService(producerID uint, pageType string, req pagetemplatedtos.RequestSavePage) (pagetemplatedtos.ResponsePageTemplate, error) {
	if err := req.Validate(); err != nil {
		return pagetemplatedtos.ResponsePageTemplate{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if !pagetemplatedtos.IsValidPageType(pageType) {
		return pagetemplatedtos.ResponsePageTemplate{}, fiber.NewError(fiber.StatusBadRequest, pagetemplatedtos.ErrPageTypeInvalid.Error())
	}
	if _, err := s.producersSvc.GetProducerByIDService(producerID); err != nil {
		return pagetemplatedtos.ResponsePageTemplate{}, err
	}

	saved, err := s.repo.UpsertRepository(producerID, pageType, req.PuckJSON)
	if err != nil {
		return pagetemplatedtos.ResponsePageTemplate{}, fiber.NewError(fiber.StatusInternalServerError, "error al guardar la página")
	}
	var response pagetemplatedtos.ResponsePageTemplate
	response.FromEntity(saved)
	return response, nil
}
