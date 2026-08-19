package page_templates

import (
	"encoding/json"
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/internal/database"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"gorm.io/gorm"
)

type Repository interface {
	GetByProducerRepository(producerID uint) ([]entities.PageTemplate, error)
	GetOneRepository(producerID uint, pageType string) (entities.PageTemplate, error)
	UpsertRepository(producerID uint, pageType string, puckJSON json.RawMessage) (entities.PageTemplate, error)
}

type repository struct {
	db *database.PostgresClient
}

func NewPageTemplatesRepository(db *database.PostgresClient) Repository {
	return &repository{db: db}
}

func (r *repository) GetByProducerRepository(producerID uint) ([]entities.PageTemplate, error) {
	var list []entities.PageTemplate
	if err := r.db.Where("producer_id = ?", producerID).
		Order("page_type ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *repository) GetOneRepository(producerID uint, pageType string) (entities.PageTemplate, error) {
	var entity entities.PageTemplate
	if err := r.db.Where("producer_id = ? AND page_type = ?", producerID, pageType).
		First(&entity).Error; err != nil {
		return entities.PageTemplate{}, err
	}
	return entity, nil
}

func (r *repository) UpsertRepository(producerID uint, pageType string, puckJSON json.RawMessage) (entities.PageTemplate, error) {
	var existing entities.PageTemplate
	err := r.db.Where("producer_id = ? AND page_type = ?", producerID, pageType).First(&existing).Error
	if err == nil {
		if err := r.db.Model(&existing).Updates(map[string]interface{}{"puck_json": puckJSON}).Error; err != nil {
			return entities.PageTemplate{}, err
		}
		return r.GetOneRepository(producerID, pageType)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return entities.PageTemplate{}, err
	}

	created := entities.PageTemplate{
		ProducerID: producerID,
		PageType:   pageType,
		PuckJSON:   puckJSON,
	}
	if err := r.db.Create(&created).Error; err != nil {
		return entities.PageTemplate{}, err
	}
	return r.GetOneRepository(producerID, pageType)
}
