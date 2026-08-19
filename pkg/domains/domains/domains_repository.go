package domains

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/internal/database"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
)

type Repository interface {
	CreateRepository(entity entities.Domain) (entities.Domain, error)
	GetByProducerRepository(producerID uint) ([]entities.Domain, error)
	GetOneRepository(filter filters.DomainFilter) (entities.Domain, error)
	DeleteRepository(id uint) error
}

type repository struct {
	db *database.PostgresClient
}

func NewDomainsRepository(db *database.PostgresClient) Repository {
	return &repository{db: db}
}

func (r *repository) CreateRepository(entity entities.Domain) (entities.Domain, error) {
	if err := r.db.Create(&entity).Error; err != nil {
		return entities.Domain{}, err
	}
	return entity, nil
}

func (r *repository) GetByProducerRepository(producerID uint) ([]entities.Domain, error) {
	var list []entities.Domain
	if err := r.db.Where("producer_id = ?", producerID).
		Order("type ASC, domain ASC").
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

func (r *repository) GetOneRepository(filter filters.DomainFilter) (entities.Domain, error) {
	query := r.db.Model(&entities.Domain{})
	if filter.ID > 0 {
		query = query.Where("id = ?", filter.ID)
	}
	if filter.ProducerID > 0 {
		query = query.Where("producer_id = ?", filter.ProducerID)
	}
	if filter.Domain != "" {
		query = query.Where("domain = ?", filter.Domain)
	}
	var entity entities.Domain
	if err := query.First(&entity).Error; err != nil {
		return entities.Domain{}, err
	}
	return entity, nil
}

func (r *repository) DeleteRepository(id uint) error {
	return r.db.Delete(&entities.Domain{}, id).Error
}
