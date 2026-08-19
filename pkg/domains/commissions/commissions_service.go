package commissions

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producers"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/commissiondtos"
	"github.com/gofiber/fiber/v2"
)

type Service interface {
	ListByProducerService(producerID uint) (commissiondtos.ResponseCommissions, error)
	CreateService(producerID uint, req commissiondtos.RequestCreateCommission) (commissiondtos.ResponseCommission, error)
}

type service struct {
	repo         Repository
	producersSvc producers.Service
}

func NewCommissionsService(repo Repository, producersSvc producers.Service) Service {
	return &service{repo: repo, producersSvc: producersSvc}
}

func (s *service) ListByProducerService(producerID uint) (commissiondtos.ResponseCommissions, error) {
	if _, err := s.producersSvc.GetProducerByIDService(producerID); err != nil {
		return commissiondtos.ResponseCommissions{}, err
	}
	list, err := s.repo.GetByProducerRepository(producerID)
	if err != nil {
		return commissiondtos.ResponseCommissions{}, fiber.NewError(fiber.StatusInternalServerError, "error al obtener comisiones")
	}
	var response commissiondtos.ResponseCommissions
	response.FromEntities(list)
	return response, nil
}

func (s *service) CreateService(producerID uint, req commissiondtos.RequestCreateCommission) (commissiondtos.ResponseCommission, error) {
	if err := req.Validate(); err != nil {
		return commissiondtos.ResponseCommission{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if _, err := s.producersSvc.GetProducerByIDService(producerID); err != nil {
		return commissiondtos.ResponseCommission{}, err
	}
	created, err := s.repo.CreateWithCloseRepository(producerID, req.Percentage, req.ValidFrom)
	if err != nil {
		return commissiondtos.ResponseCommission{}, fiber.NewError(fiber.StatusInternalServerError, "error al registrar comisión")
	}
	var response commissiondtos.ResponseCommission
	response.FromEntity(created)
	return response, nil
}
