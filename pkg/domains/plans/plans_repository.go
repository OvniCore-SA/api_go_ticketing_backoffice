package plans

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/internal/database"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"gorm.io/gorm"
)

type Repository interface {
	CreatePlanRepository(entity entities.Plan) (entities.Plan, error)
	GetAllPlansRepository(filter filters.PlanFilter, page, limit int) ([]entities.Plan, int64, error)
	GetPlanRepository(filter filters.PlanFilter) (entities.Plan, error)
	UpdatePlanRepository(id uint, fields map[string]interface{}) error
	DeletePlanRepository(id uint) error
	CountProducersByPlanRepository(planID uint) (int64, error)
}

type repository struct {
	db *database.PostgresClient
}

func NewPlansRepository(db *database.PostgresClient) Repository {
	return &repository{db: db}
}

func (r *repository) CreatePlanRepository(entity entities.Plan) (entities.Plan, error) {
	if err := r.db.Create(&entity).Error; err != nil {
		return entities.Plan{}, err
	}
	return entity, nil
}

func (r *repository) GetAllPlansRepository(filter filters.PlanFilter, page, limit int) ([]entities.Plan, int64, error) {
	var list []entities.Plan
	var total int64

	query := r.db.Model(&entities.Plan{})

	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ?", like, like)
	}
	if filter.IsActive != nil {
		query = query.Where("is_active = ?", *filter.IsActive)
	}

	query.Count(&total)

	if err := query.Order("name ASC").
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&list).Error; err != nil {
		return nil, 0, err
	}
	return list, total, nil
}

func (r *repository) GetPlanRepository(filter filters.PlanFilter) (entities.Plan, error) {
	query := r.db.Model(&entities.Plan{})
	if filter.ID > 0 {
		query = query.Where("id = ?", filter.ID)
	}
	if filter.Code != "" {
		query = query.Where("code = ?", filter.Code)
	}
	if filter.ID == 0 && filter.Code == "" {
		return entities.Plan{}, errors.New("se requiere id o code")
	}

	var entity entities.Plan
	if err := query.First(&entity).Error; err != nil {
		return entities.Plan{}, err
	}
	return entity, nil
}

func (r *repository) UpdatePlanRepository(id uint, fields map[string]interface{}) error {
	return r.db.Model(&entities.Plan{}).Where("id = ?", id).Updates(fields).Error
}

func (r *repository) DeletePlanRepository(id uint) error {
	return r.db.Delete(&entities.Plan{}, id).Error
}

func (r *repository) CountProducersByPlanRepository(planID uint) (int64, error) {
	var count int64
	err := r.db.Model(&entities.Producer{}).
		Where("plan_id = ? AND deleted_at IS NULL", planID).
		Count(&count).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, err
	}
	return count, nil
}
