package producer_component_variants

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/internal/database"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"gorm.io/gorm"
)

type Repository interface {
	GetByProducerRepository(producerID uint) ([]entities.ProducerComponentVariant, error)
	GetOneRepository(producerID, moduleID uint) (entities.ProducerComponentVariant, error)
	UpsertRepository(producerID, moduleID, componentVariantID uint) (entities.ProducerComponentVariant, error)
}

type repository struct {
	db *database.PostgresClient
}

func NewProducerComponentVariantsRepository(db *database.PostgresClient) Repository {
	return &repository{db: db}
}

func (r *repository) GetByProducerRepository(producerID uint) ([]entities.ProducerComponentVariant, error) {
	var list []entities.ProducerComponentVariant
	if err := r.db.Preload("ComponentVariant").
		Where("producer_id = ?", producerID).
		Order("module_id ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *repository) GetOneRepository(producerID, moduleID uint) (entities.ProducerComponentVariant, error) {
	var entity entities.ProducerComponentVariant
	if err := r.db.Preload("ComponentVariant").
		Where("producer_id = ? AND module_id = ?", producerID, moduleID).
		First(&entity).Error; err != nil {
		return entities.ProducerComponentVariant{}, err
	}
	return entity, nil
}

func (r *repository) UpsertRepository(producerID, moduleID, componentVariantID uint) (entities.ProducerComponentVariant, error) {
	var existing entities.ProducerComponentVariant
	err := r.db.Where("producer_id = ? AND module_id = ?", producerID, moduleID).
		First(&existing).Error
	if err == nil {
		if err := r.db.Model(&existing).Updates(map[string]interface{}{
			"component_variant_id": componentVariantID,
		}).Error; err != nil {
			return entities.ProducerComponentVariant{}, err
		}
		return r.GetOneRepository(producerID, moduleID)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return entities.ProducerComponentVariant{}, err
	}

	created := entities.ProducerComponentVariant{
		ProducerID:         producerID,
		ModuleID:           moduleID,
		ComponentVariantID: componentVariantID,
	}
	if err := r.db.Create(&created).Error; err != nil {
		return entities.ProducerComponentVariant{}, err
	}
	return r.GetOneRepository(producerID, moduleID)
}
