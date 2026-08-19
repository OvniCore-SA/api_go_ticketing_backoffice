package domains

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producers"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/domaindtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Service interface {
	ListByProducerService(producerID uint) (domaindtos.ResponseDomains, error)
	CreateService(producerID uint, req domaindtos.RequestCreateDomain) (domaindtos.ResponseDomain, error)
	DeleteService(producerID, id uint) error
}

type service struct {
	repo         Repository
	producersSvc producers.Service
}

func NewDomainsService(repo Repository, producersSvc producers.Service) Service {
	return &service{repo: repo, producersSvc: producersSvc}
}

func (s *service) ListByProducerService(producerID uint) (domaindtos.ResponseDomains, error) {
	if _, err := s.producersSvc.GetProducerByIDService(producerID); err != nil {
		return domaindtos.ResponseDomains{}, err
	}
	list, err := s.repo.GetByProducerRepository(producerID)
	if err != nil {
		return domaindtos.ResponseDomains{}, fiber.NewError(fiber.StatusInternalServerError, "error al obtener dominios")
	}
	var response domaindtos.ResponseDomains
	response.FromEntities(list)
	return response, nil
}

func (s *service) CreateService(producerID uint, req domaindtos.RequestCreateDomain) (domaindtos.ResponseDomain, error) {
	if err := req.Validate(); err != nil {
		return domaindtos.ResponseDomain{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if _, err := s.producersSvc.GetProducerByIDService(producerID); err != nil {
		return domaindtos.ResponseDomain{}, err
	}
	if _, err := s.repo.GetOneRepository(filters.DomainFilter{Domain: req.Domain}); err == nil {
		return domaindtos.ResponseDomain{}, fiber.NewError(fiber.StatusConflict, domaindtos.ErrDomainExists.Error())
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return domaindtos.ResponseDomain{}, fiber.NewError(fiber.StatusInternalServerError, "error al validar dominio")
	}

	created, err := s.repo.CreateRepository(req.ToEntity(producerID))
	if err != nil {
		return domaindtos.ResponseDomain{}, fiber.NewError(fiber.StatusInternalServerError, "error al registrar dominio")
	}
	var response domaindtos.ResponseDomain
	response.FromEntity(created)
	return response, nil
}

func (s *service) DeleteService(producerID, id uint) error {
	entity, err := s.repo.GetOneRepository(filters.DomainFilter{ID: id, ProducerID: producerID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, domaindtos.ErrDomainNotFound.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "error al obtener dominio")
	}
	if err := s.repo.DeleteRepository(entity.ID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error al eliminar dominio")
	}
	return nil
}
