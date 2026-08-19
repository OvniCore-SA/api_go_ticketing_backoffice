package producer_modules

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/modules"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producers"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/producermoduledtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Service interface {
	ListByProducerService(producerID uint) (producermoduledtos.ResponseProducerModules, error)
	ToggleModuleService(producerID, moduleID uint, req producermoduledtos.RequestToggleModule) (producermoduledtos.ResponseProducerModule, error)
}

type service struct {
	repo         Repository
	modulesRepo  modules.Repository
	producersSvc producers.Service
}

func NewProducerModulesService(repo Repository, modulesRepo modules.Repository, producersSvc producers.Service) Service {
	return &service{repo: repo, modulesRepo: modulesRepo, producersSvc: producersSvc}
}

func (s *service) ListByProducerService(producerID uint) (producermoduledtos.ResponseProducerModules, error) {
	if _, err := s.producersSvc.GetProducerByIDService(producerID); err != nil {
		return producermoduledtos.ResponseProducerModules{}, err
	}
	list, err := s.repo.GetByProducerRepository(producerID)
	if err != nil {
		return producermoduledtos.ResponseProducerModules{}, fiber.NewError(fiber.StatusInternalServerError, "error al obtener módulos del producer")
	}
	var response producermoduledtos.ResponseProducerModules
	response.FromEntities(list)
	return response, nil
}

func (s *service) ToggleModuleService(producerID, moduleID uint, req producermoduledtos.RequestToggleModule) (producermoduledtos.ResponseProducerModule, error) {
	if err := req.Validate(); err != nil {
		return producermoduledtos.ResponseProducerModule{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if _, err := s.producersSvc.GetProducerByIDService(producerID); err != nil {
		return producermoduledtos.ResponseProducerModule{}, err
	}
	module, err := s.modulesRepo.GetModuleRepository(filters.ModuleFilter{ID: moduleID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return producermoduledtos.ResponseProducerModule{}, fiber.NewError(fiber.StatusNotFound, "módulo no encontrado")
		}
		return producermoduledtos.ResponseProducerModule{}, fiber.NewError(fiber.StatusInternalServerError, "error al validar módulo")
	}

	// Regla de negocio: los módulos core no se pueden deshabilitar.
	if module.IsCore && !*req.Enabled {
		return producermoduledtos.ResponseProducerModule{}, fiber.NewError(fiber.StatusConflict, producermoduledtos.ErrCoreModuleDisallowed.Error())
	}

	saved, err := s.repo.UpsertRepository(producerID, moduleID, *req.Enabled)
	if err != nil {
		return producermoduledtos.ResponseProducerModule{}, fiber.NewError(fiber.StatusInternalServerError, "error al actualizar módulo del producer")
	}
	var response producermoduledtos.ResponseProducerModule
	response.FromEntity(saved)
	return response, nil
}
