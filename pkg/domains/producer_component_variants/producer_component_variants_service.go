package producer_component_variants

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/modules"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producers"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/producervariantdtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Service interface {
	ListByProducerService(producerID uint) (producervariantdtos.ResponseProducerComponentVariants, error)
	AssignVariantService(producerID, moduleID uint, req producervariantdtos.RequestAssignVariant) (producervariantdtos.ResponseProducerComponentVariant, error)
}

type service struct {
	repo         Repository
	modulesRepo  modules.Repository
	producersSvc producers.Service
}

func NewProducerComponentVariantsService(repo Repository, modulesRepo modules.Repository, producersSvc producers.Service) Service {
	return &service{repo: repo, modulesRepo: modulesRepo, producersSvc: producersSvc}
}

func (s *service) ListByProducerService(producerID uint) (producervariantdtos.ResponseProducerComponentVariants, error) {
	if _, err := s.producersSvc.GetProducerByIDService(producerID); err != nil {
		return producervariantdtos.ResponseProducerComponentVariants{}, err
	}
	list, err := s.repo.GetByProducerRepository(producerID)
	if err != nil {
		return producervariantdtos.ResponseProducerComponentVariants{}, fiber.NewError(fiber.StatusInternalServerError, "error al obtener variantes del producer")
	}
	var response producervariantdtos.ResponseProducerComponentVariants
	response.FromEntities(list)
	return response, nil
}

func (s *service) AssignVariantService(producerID, moduleID uint, req producervariantdtos.RequestAssignVariant) (producervariantdtos.ResponseProducerComponentVariant, error) {
	if err := req.Validate(); err != nil {
		return producervariantdtos.ResponseProducerComponentVariant{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if _, err := s.producersSvc.GetProducerByIDService(producerID); err != nil {
		return producervariantdtos.ResponseProducerComponentVariant{}, err
	}
	if _, err := s.modulesRepo.GetModuleRepository(filters.ModuleFilter{ID: moduleID}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return producervariantdtos.ResponseProducerComponentVariant{}, fiber.NewError(fiber.StatusNotFound, "módulo no encontrado")
		}
		return producervariantdtos.ResponseProducerComponentVariant{}, fiber.NewError(fiber.StatusInternalServerError, "error al validar módulo")
	}
	// Validar que la variante pertenezca al módulo indicado.
	variant, err := s.modulesRepo.GetVariantRepository(filters.ComponentVariantFilter{ID: req.ComponentVariantID})
	if err != nil {
		return producervariantdtos.ResponseProducerComponentVariant{}, fiber.NewError(fiber.StatusBadRequest, "variante no encontrada")
	}
	if variant.ModuleID != moduleID {
		return producervariantdtos.ResponseProducerComponentVariant{}, fiber.NewError(fiber.StatusBadRequest, producervariantdtos.ErrVariantNotInModule.Error())
	}

	saved, err := s.repo.UpsertRepository(producerID, moduleID, req.ComponentVariantID)
	if err != nil {
		return producervariantdtos.ResponseProducerComponentVariant{}, fiber.NewError(fiber.StatusInternalServerError, "error al asignar variante")
	}
	var response producervariantdtos.ResponseProducerComponentVariant
	response.FromEntity(saved)
	return response, nil
}
