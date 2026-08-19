package commissions

import (
	"time"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/internal/database"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"gorm.io/gorm"
)

type Repository interface {
	GetByProducerRepository(producerID uint) ([]entities.Commission, error)
	CreateWithCloseRepository(producerID uint, percentage float64, validFrom time.Time) (entities.Commission, error)
}

type repository struct {
	db *database.PostgresClient
}

func NewCommissionsRepository(db *database.PostgresClient) Repository {
	return &repository{db: db}
}

func (r *repository) GetByProducerRepository(producerID uint) ([]entities.Commission, error) {
	var list []entities.Commission
	if err := r.db.Where("producer_id = ?", producerID).
		Order("valid_from DESC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// CreateWithCloseRepository cierra la comisión vigente (valid_to NULL) del
// producer con la nueva ValidFrom y crea el nuevo registro, todo en una
// transacción. Así siempre hay a lo sumo una vigente por producer.
func (r *repository) CreateWithCloseRepository(producerID uint, percentage float64, validFrom time.Time) (entities.Commission, error) {
	var created entities.Commission
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&entities.Commission{}).
			Where("producer_id = ? AND valid_to IS NULL", producerID).
			Update("valid_to", validFrom).Error; err != nil {
			return err
		}
		created = entities.Commission{
			ProducerID: producerID,
			Percentage: percentage,
			ValidFrom:  validFrom,
		}
		return tx.Create(&created).Error
	})
	if err != nil {
		return entities.Commission{}, err
	}
	return created, nil
}
