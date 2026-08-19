package producers

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/internal/database"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"gorm.io/gorm"
)

type Repository interface {
	CreateProducerRepository(entity entities.Producer) (entities.Producer, error)
	GetAllProducersRepository(filter filters.ProducerFilter, page, limit int) ([]entities.Producer, int64, error)
	GetProducerRepository(filter filters.ProducerFilter) (entities.Producer, error)
	UpdateProducerRepository(id uint, fields map[string]interface{}) error
	DeleteProducerRepository(id uint) error

	// SeedFromTemplateRepository crea el producer y, si tiene TemplateID,
	// copia módulos + páginas + design tokens dentro de la MISMA transacción.
	// Si algo falla, no queda ningún registro persistido.
	SeedProducerFromTemplateRepository(entity entities.Producer, templateID *uint) (entities.Producer, error)
}

type repository struct {
	db *database.PostgresClient
}

func NewProducersRepository(db *database.PostgresClient) Repository {
	return &repository{db: db}
}

func (r *repository) CreateProducerRepository(entity entities.Producer) (entities.Producer, error) {
	if err := r.db.Create(&entity).Error; err != nil {
		return entities.Producer{}, err
	}
	return entity, nil
}

func (r *repository) GetAllProducersRepository(filter filters.ProducerFilter, page, limit int) ([]entities.Producer, int64, error) {
	var list []entities.Producer
	var total int64

	query := r.db.Model(&entities.Producer{})
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ? OR slug ILIKE ? OR contact_email ILIKE ?", like, like, like)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.PlanID > 0 {
		query = query.Where("plan_id = ?", filter.PlanID)
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

func (r *repository) GetProducerRepository(filter filters.ProducerFilter) (entities.Producer, error) {
	// Preload Template + Plan siempre — es la lectura más caliente del sistema
	// (la usa el Resolver de producer_config en cada render) y la traemos en
	// un solo query con LEFT JOIN.
	query := r.db.Model(&entities.Producer{}).
		Preload("Template").
		Preload("Plan")
	if filter.ID > 0 {
		query = query.Where("producers.id = ?", filter.ID)
	}
	if filter.Slug != "" {
		query = query.Where("producers.slug = ?", filter.Slug)
	}
	if filter.ID == 0 && filter.Slug == "" {
		return entities.Producer{}, errors.New("se requiere id o slug")
	}
	var entity entities.Producer
	if err := query.First(&entity).Error; err != nil {
		return entities.Producer{}, err
	}
	return entity, nil
}

func (r *repository) UpdateProducerRepository(id uint, fields map[string]interface{}) error {
	return r.db.Model(&entities.Producer{}).Where("id = ?", id).Updates(fields).Error
}

func (r *repository) DeleteProducerRepository(id uint) error {
	return r.db.Delete(&entities.Producer{}, id).Error
}

// SeedProducerFromTemplateRepository es un create-with-side-effects transaccional:
//  1. Crea el Producer.
//  2. Si tiene TemplateID válido, lee TemplateModule/TemplatePage/default_*
//     y los copia como ProducerModule/PageTemplate/DesignTokens.
//  3. Si no tiene template, crea un DesignTokens vacío (para tener siempre 1:1).
func (r *repository) SeedProducerFromTemplateRepository(entity entities.Producer, templateID *uint) (entities.Producer, error) {
	err := r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&entity).Error; err != nil {
			return err
		}

		if templateID == nil || *templateID == 0 {
			// Sin template: dejar DesignTokens vacío para tener siempre 1:1.
			tokens := entities.DesignTokens{ProducerID: entity.ID}
			return tx.Create(&tokens).Error
		}

		var template entities.Template
		if err := tx.Where("id = ?", *templateID).First(&template).Error; err != nil {
			return err
		}

		// Módulos: copiar TemplateModule → ProducerModule (enabled=true).
		var tmods []entities.TemplateModule
		if err := tx.Where("template_id = ?", template.ID).Find(&tmods).Error; err != nil {
			return err
		}
		if len(tmods) > 0 {
			pmods := make([]entities.ProducerModule, len(tmods))
			for i, tm := range tmods {
				pmods[i] = entities.ProducerModule{
					ProducerID: entity.ID,
					ModuleID:   tm.ModuleID,
					Enabled:    true,
				}
			}
			if err := tx.Create(&pmods).Error; err != nil {
				return err
			}
		}

		// Páginas: copiar TemplatePage → PageTemplate (puck_json_default → puck_json).
		var tpages []entities.TemplatePage
		if err := tx.Where("template_id = ?", template.ID).Find(&tpages).Error; err != nil {
			return err
		}
		if len(tpages) > 0 {
			ppages := make([]entities.PageTemplate, len(tpages))
			for i, tp := range tpages {
				ppages[i] = entities.PageTemplate{
					ProducerID: entity.ID,
					PageType:   tp.PageType,
					PuckJSON:   tp.PuckJSONDefault,
				}
			}
			if err := tx.Create(&ppages).Error; err != nil {
				return err
			}
		}

		// Design tokens: 1 registro con los defaults de la template.
		tokens := entities.DesignTokens{
			ProducerID: entity.ID,
			Colors:     template.DefaultColors,
			Fonts:      template.DefaultFonts,
			Radius:     template.DefaultRadius,
			Shadows:    template.DefaultShadows,
		}
		if err := tx.Create(&tokens).Error; err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return entities.Producer{}, err
	}
	return entity, nil
}
