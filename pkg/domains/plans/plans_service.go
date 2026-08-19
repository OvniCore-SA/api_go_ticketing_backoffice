package plans

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/plandtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/utils"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Service interface {
	CreatePlanService(req plandtos.RequestCreatePlan) (plandtos.ResponsePlan, error)
	GetPlansService(req plandtos.RequestListPlans) (plandtos.ResponsePlans, utils.Pagination, error)
	GetPlanByIDService(id uint) (plandtos.ResponsePlan, error)
	UpdatePlanService(id uint, req plandtos.RequestUpdatePlan) (plandtos.ResponsePlan, error)
	DeletePlanService(id uint) error
}

type service struct {
	repo Repository
}

func NewPlansService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) CreatePlanService(req plandtos.RequestCreatePlan) (plandtos.ResponsePlan, error) {
	if err := req.Validate(); err != nil {
		return plandtos.ResponsePlan{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if _, err := s.repo.GetPlanRepository(filters.PlanFilter{Code: req.Code}); err == nil {
		return plandtos.ResponsePlan{}, fiber.NewError(fiber.StatusConflict, plandtos.ErrPlanCodeExists.Error())
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return plandtos.ResponsePlan{}, fiber.NewError(fiber.StatusInternalServerError, "error al validar código del plan")
	}

	created, err := s.repo.CreatePlanRepository(req.ToEntity())
	if err != nil {
		return plandtos.ResponsePlan{}, fiber.NewError(fiber.StatusInternalServerError, "error al crear el plan")
	}

	var response plandtos.ResponsePlan
	response.FromEntity(created)
	return response, nil
}

func (s *service) GetPlansService(req plandtos.RequestListPlans) (response plandtos.ResponsePlans, pagination utils.Pagination, err error) {
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 {
		req.Limit = 20
	}

	filter := filters.PlanFilter{Search: req.Search, IsActive: req.IsActive}
	list, total, err := s.repo.GetAllPlansRepository(filter, req.Page, req.Limit)
	if err != nil {
		err = fiber.NewError(fiber.StatusInternalServerError, "error al obtener planes")
		return
	}
	response.FromEntities(list)
	pagination = utils.NewPagination(req.Page, req.Limit, total)
	return
}

func (s *service) GetPlanByIDService(id uint) (plandtos.ResponsePlan, error) {
	entity, err := s.repo.GetPlanRepository(filters.PlanFilter{ID: id})
	if err != nil {
		return plandtos.ResponsePlan{}, fiber.NewError(fiber.StatusNotFound, plandtos.ErrPlanNotFound.Error())
	}
	var response plandtos.ResponsePlan
	response.FromEntity(entity)
	return response, nil
}

func (s *service) UpdatePlanService(id uint, req plandtos.RequestUpdatePlan) (plandtos.ResponsePlan, error) {
	if _, err := s.repo.GetPlanRepository(filters.PlanFilter{ID: id}); err != nil {
		return plandtos.ResponsePlan{}, fiber.NewError(fiber.StatusNotFound, plandtos.ErrPlanNotFound.Error())
	}

	fields := map[string]interface{}{}
	if req.Name != "" {
		fields["name"] = req.Name
	}
	if req.Description != "" {
		fields["description"] = req.Description
	}
	if req.IsActive != nil {
		fields["is_active"] = *req.IsActive
	}

	if len(fields) > 0 {
		if err := s.repo.UpdatePlanRepository(id, fields); err != nil {
			return plandtos.ResponsePlan{}, fiber.NewError(fiber.StatusInternalServerError, "error al actualizar el plan")
		}
	}

	updated, _ := s.repo.GetPlanRepository(filters.PlanFilter{ID: id})
	var response plandtos.ResponsePlan
	response.FromEntity(updated)
	return response, nil
}

func (s *service) DeletePlanService(id uint) error {
	if _, err := s.repo.GetPlanRepository(filters.PlanFilter{ID: id}); err != nil {
		return fiber.NewError(fiber.StatusNotFound, plandtos.ErrPlanNotFound.Error())
	}

	count, err := s.repo.CountProducersByPlanRepository(id)
	if err == nil && count > 0 {
		return fiber.NewError(fiber.StatusConflict, plandtos.ErrPlanHasProducers.Error())
	}

	if err := s.repo.DeletePlanRepository(id); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "error al eliminar el plan")
	}
	return nil
}
