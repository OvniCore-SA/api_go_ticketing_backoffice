package producer_modules

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/internal/database"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"gorm.io/gorm"
)

type Repository interface {
	GetByProducerRepository(producerID uint) ([]entities.ProducerModule, error)
	GetOneRepository(producerID, moduleID uint) (entities.ProducerModule, error)
	UpsertRepository(producerID, moduleID uint, enabled bool) (entities.ProducerModule, error)
}

type repository struct {
	db *database.PostgresClient
}

func NewProducerModulesRepository(db *database.PostgresClient) Repository {
	return &repository{db: db}
}

func (r *repository) GetByProducerRepository(producerID uint) ([]entities.ProducerModule, error) {
	var list []entities.ProducerModule
	if err := r.db.Preload("Module").
		Where("producer_id = ?", producerID).
		Order("module_id ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *repository) GetOneRepository(producerID, moduleID uint) (entities.ProducerModule, error) {
	var entity entities.ProducerModule
	if err := r.db.Preload("Module").
		Where("producer_id = ? AND module_id = ?", producerID, moduleID).
		First(&entity).Error; err != nil {
		return entities.ProducerModule{}, err
	}
	return entity, nil
}

func (r *repository) UpsertRepository(producerID, moduleID uint, enabled bool) (entities.ProducerModule, error) {
	var existing entities.ProducerModule
	err := r.db.Where("producer_id = ? AND module_id = ?", producerID, moduleID).
		First(&existing).Error
	if err == nil {
		if err := r.db.Model(&existing).Updates(map[string]interface{}{"enabled": enabled}).Error; err != nil {
			return entities.ProducerModule{}, err
		}
		return r.GetOneRepository(producerID, moduleID)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return entities.ProducerModule{}, err
	}

	created := entities.ProducerModule{ProducerID: producerID, ModuleID: moduleID, Enabled: enabled}
	if err := r.db.Create(&created).Error; err != nil {
		return entities.ProducerModule{}, err
	}
	return r.GetOneRepository(producerID, moduleID)
}
