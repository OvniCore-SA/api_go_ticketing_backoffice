package design_tokens

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producers"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/designtokendtos"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Service interface {
	GetTokensService(producerID uint) (designtokendtos.ResponseDesignTokens, error)
	UpdateTokensService(producerID uint, req designtokendtos.RequestUpdateDesignTokens) (designtokendtos.ResponseDesignTokens, error)
}

type service struct {
	repo         Repository
	producersSvc producers.Service
}

func NewDesignTokensService(repo Repository, producersSvc producers.Service) Service {
	return &service{repo: repo, producersSvc: producersSvc}
}

func (s *service) GetTokensService(producerID uint) (designtokendtos.ResponseDesignTokens, error) {
	if _, err := s.producersSvc.GetProducerByIDService(producerID); err != nil {
		return designtokendtos.ResponseDesignTokens{}, err
	}
	entity, err := s.repo.GetByProducerRepository(producerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return designtokendtos.ResponseDesignTokens{}, fiber.NewError(fiber.StatusNotFound, designtokendtos.ErrTokensNotFound.Error())
		}
		return designtokendtos.ResponseDesignTokens{}, fiber.NewError(fiber.StatusInternalServerError, "error al obtener design tokens")
	}
	var response designtokendtos.ResponseDesignTokens
	response.FromEntity(entity)
	return response, nil
}

func (s *service) UpdateTokensService(producerID uint, req designtokendtos.RequestUpdateDesignTokens) (designtokendtos.ResponseDesignTokens, error) {
	if err := req.Validate(); err != nil {
		return designtokendtos.ResponseDesignTokens{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if _, err := s.producersSvc.GetProducerByIDService(producerID); err != nil {
		return designtokendtos.ResponseDesignTokens{}, err
	}

	fields := map[string]interface{}{}
	if len(req.Colors) > 0 {
		fields["colors"] = req.Colors
	}
	if len(req.Fonts) > 0 {
		fields["fonts"] = req.Fonts
	}
	if len(req.Radius) > 0 {
		fields["radius"] = req.Radius
	}
	if len(req.Shadows) > 0 {
		fields["shadows"] = req.Shadows
	}

	saved, err := s.repo.UpsertRepository(producerID, fields)
	if err != nil {
		return designtokendtos.ResponseDesignTokens{}, fiber.NewError(fiber.StatusInternalServerError, "error al actualizar design tokens")
	}
	var response designtokendtos.ResponseDesignTokens
	response.FromEntity(saved)
	return response, nil
}
