package modules

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/moduledtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"github.com/gofiber/fiber/v2"
)

type Service interface {
	GetModulesService(req moduledtos.RequestListModules) (moduledtos.ResponseModules, error)
	GetModuleByIDService(id uint) (moduledtos.ResponseModule, error)
	GetVariantsByModuleService(moduleID uint) (moduledtos.ResponseComponentVariants, error)
}

type service struct {
	repo Repository
}

func NewModulesService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetModulesService(req moduledtos.RequestListModules) (moduledtos.ResponseModules, error) {
	list, err := s.repo.GetAllModulesRepository(filters.ModuleFilter{
		Search:   req.Search,
		Category: req.Category,
		IsCore:   req.IsCore,
	})
	if err != nil {
		return moduledtos.ResponseModules{}, fiber.NewError(fiber.StatusInternalServerError, "error al obtener módulos")
	}
	var response moduledtos.ResponseModules
	response.FromEntities(list)
	return response, nil
}

func (s *service) GetModuleByIDService(id uint) (moduledtos.ResponseModule, error) {
	entity, err := s.repo.GetModuleRepository(filters.ModuleFilter{ID: id})
	if err != nil {
		return moduledtos.ResponseModule{}, fiber.NewError(fiber.StatusNotFound, moduledtos.ErrModuleNotFound.Error())
	}
	var response moduledtos.ResponseModule
	response.FromEntity(entity)
	return response, nil
}

func (s *service) GetVariantsByModuleService(moduleID uint) (moduledtos.ResponseComponentVariants, error) {
	if _, err := s.repo.GetModuleRepository(filters.ModuleFilter{ID: moduleID}); err != nil {
		return moduledtos.ResponseComponentVariants{}, fiber.NewError(fiber.StatusNotFound, moduledtos.ErrModuleNotFound.Error())
	}
	list, err := s.repo.GetVariantsByModuleRepository(moduleID)
	if err != nil {
		return moduledtos.ResponseComponentVariants{}, fiber.NewError(fiber.StatusInternalServerError, "error al obtener variantes")
	}
	var response moduledtos.ResponseComponentVariants
	response.FromEntities(list)
	return response, nil
}
