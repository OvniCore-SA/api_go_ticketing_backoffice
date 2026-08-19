package modules

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/internal/database"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
)

type Repository interface {
	GetAllModulesRepository(filter filters.ModuleFilter) ([]entities.Module, error)
	GetModuleRepository(filter filters.ModuleFilter) (entities.Module, error)
	GetVariantsByModuleRepository(moduleID uint) ([]entities.ComponentVariant, error)
	GetVariantRepository(filter filters.ComponentVariantFilter) (entities.ComponentVariant, error)
}

type repository struct {
	db *database.PostgresClient
}

func NewModulesRepository(db *database.PostgresClient) Repository {
	return &repository{db: db}
}

func (r *repository) GetAllModulesRepository(filter filters.ModuleFilter) ([]entities.Module, error) {
	var list []entities.Module
	query := r.db.Model(&entities.Module{})

	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ?", like, like)
	}
	if filter.Category != "" {
		query = query.Where("category = ?", filter.Category)
	}
	if filter.IsCore != nil {
		query = query.Where("is_core = ?", *filter.IsCore)
	}

	if err := query.Order("code ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *repository) GetModuleRepository(filter filters.ModuleFilter) (entities.Module, error) {
	query := r.db.Model(&entities.Module{})
	if filter.ID > 0 {
		query = query.Where("id = ?", filter.ID)
	}
	if filter.Code != "" {
		query = query.Where("code = ?", filter.Code)
	}
	if filter.ID == 0 && filter.Code == "" {
		return entities.Module{}, errors.New("se requiere id o code")
	}

	var entity entities.Module
	if err := query.First(&entity).Error; err != nil {
		return entities.Module{}, err
	}
	return entity, nil
}

func (r *repository) GetVariantsByModuleRepository(moduleID uint) ([]entities.ComponentVariant, error) {
	var list []entities.ComponentVariant
	if err := r.db.
		Where("module_id = ?", moduleID).
		Order("code ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *repository) GetVariantRepository(filter filters.ComponentVariantFilter) (entities.ComponentVariant, error) {
	query := r.db.Model(&entities.ComponentVariant{})
	if filter.ID > 0 {
		query = query.Where("id = ?", filter.ID)
	}
	if filter.ModuleID > 0 {
		query = query.Where("module_id = ?", filter.ModuleID)
	}
	if filter.Code != "" {
		query = query.Where("code = ?", filter.Code)
	}

	var entity entities.ComponentVariant
	if err := query.First(&entity).Error; err != nil {
		return entities.ComponentVariant{}, err
	}
	return entity, nil
}
