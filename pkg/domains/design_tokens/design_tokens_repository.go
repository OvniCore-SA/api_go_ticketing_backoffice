package design_tokens

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/internal/database"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"gorm.io/gorm"
)

type Repository interface {
	GetByProducerRepository(producerID uint) (entities.DesignTokens, error)
	UpsertRepository(producerID uint, fields map[string]interface{}) (entities.DesignTokens, error)
}

type repository struct {
	db *database.PostgresClient
}

func NewDesignTokensRepository(db *database.PostgresClient) Repository {
	return &repository{db: db}
}

func (r *repository) GetByProducerRepository(producerID uint) (entities.DesignTokens, error) {
	var entity entities.DesignTokens
	if err := r.db.Where("producer_id = ?", producerID).First(&entity).Error; err != nil {
		return entities.DesignTokens{}, err
	}
	return entity, nil
}

func (r *repository) UpsertRepository(producerID uint, fields map[string]interface{}) (entities.DesignTokens, error) {
	var existing entities.DesignTokens
	err := r.db.Where("producer_id = ?", producerID).First(&existing).Error
	if err == nil {
		if len(fields) > 0 {
			if err := r.db.Model(&existing).Updates(fields).Error; err != nil {
				return entities.DesignTokens{}, err
			}
		}
		return r.GetByProducerRepository(producerID)
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return entities.DesignTokens{}, err
	}

	created := entities.DesignTokens{ProducerID: producerID}
	// Aplicar fields al struct antes de crearlo — GORM ignora las claves del map
	// en Create; usamos otro approach: crear vacío y luego update.
	if err := r.db.Create(&created).Error; err != nil {
		return entities.DesignTokens{}, err
	}
	if len(fields) > 0 {
		if err := r.db.Model(&created).Updates(fields).Error; err != nil {
			return entities.DesignTokens{}, err
		}
	}
	return r.GetByProducerRepository(producerID)
}
