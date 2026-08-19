package templates

import (
	"encoding/json"
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/internal/database"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"gorm.io/gorm"
)

type Repository interface {
	CreateTemplateRepository(entity entities.Template) (entities.Template, error)
	GetAllTemplatesRepository(filter filters.TemplateFilter, page, limit int) ([]entities.Template, int64, error)
	GetTemplateRepository(filter filters.TemplateFilter) (entities.Template, error)
	UpdateTemplateRepository(id uint, fields map[string]interface{}) error
	DeleteTemplateRepository(id uint) error

	// CountProducersByTemplateRepository devuelve cuántos Producers vivos
	// referencian esta template. Se usa como guardarraíl en el DELETE:
	// borrar una template en uso dejaría producers con `template_id`
	// colgado y al Resolver sin capa de defaults.
	CountProducersByTemplateRepository(templateID uint) (int64, error)

	GetTemplateModulesRepository(templateID uint) ([]entities.TemplateModule, error)
	ReplaceTemplateModulesRepository(templateID uint, moduleIDs []uint) error

	GetTemplatePageRepository(templateID uint, pageType string) (entities.TemplatePage, error)
	GetTemplatePagesRepository(templateID uint) ([]entities.TemplatePage, error)
	UpsertTemplatePageRepository(templateID uint, pageType string, puckJSON json.RawMessage) (entities.TemplatePage, error)
}

type repository struct {
	db *database.PostgresClient
}

func NewTemplatesRepository(db *database.PostgresClient) Repository {
	return &repository{db: db}
}

func (r *repository) CreateTemplateRepository(entity entities.Template) (entities.Template, error) {
	if err := r.db.Create(&entity).Error; err != nil {
		return entities.Template{}, err
	}
	return entity, nil
}

func (r *repository) GetAllTemplatesRepository(filter filters.TemplateFilter, page, limit int) ([]entities.Template, int64, error) {
	var list []entities.Template
	var total int64

	query := r.db.Model(&entities.Template{})
	if filter.Search != "" {
		like := "%" + filter.Search + "%"
		query = query.Where("name ILIKE ? OR code ILIKE ?", like, like)
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

func (r *repository) GetTemplateRepository(filter filters.TemplateFilter) (entities.Template, error) {
	query := r.db.Model(&entities.Template{})
	if filter.ID > 0 {
		query = query.Where("id = ?", filter.ID)
	}
	if filter.Code != "" {
		query = query.Where("code = ?", filter.Code)
	}
	if filter.ID == 0 && filter.Code == "" {
		return entities.Template{}, errors.New("se requiere id o code")
	}
	var entity entities.Template
	if err := query.First(&entity).Error; err != nil {
		return entities.Template{}, err
	}
	return entity, nil
}

func (r *repository) UpdateTemplateRepository(id uint, fields map[string]interface{}) error {
	return r.db.Model(&entities.Template{}).Where("id = ?", id).Updates(fields).Error
}

func (r *repository) DeleteTemplateRepository(id uint) error {
	return r.db.Delete(&entities.Template{}, id).Error
}

func (r *repository) CountProducersByTemplateRepository(templateID uint) (int64, error) {
	var count int64
	err := r.db.Model(&entities.Producer{}).
		Where("template_id = ? AND deleted_at IS NULL", templateID).
		Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *repository) GetTemplateModulesRepository(templateID uint) ([]entities.TemplateModule, error) {
	var list []entities.TemplateModule
	if err := r.db.Preload("Module").
		Where("template_id = ?", templateID).
		Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// ReplaceTemplateModulesRepository borra el set actual (soft delete) y crea
// el nuevo set. Todo dentro de una transacción — o queda íntegro o no cambia.
func (r *repository) ReplaceTemplateModulesRepository(templateID uint, moduleIDs []uint) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("template_id = ?", templateID).
			Delete(&entities.TemplateModule{}).Error; err != nil {
			return err
		}
		if len(moduleIDs) == 0 {
			return nil
		}

		// Validar que todos los módulos existan.
		var count int64
		if err := tx.Model(&entities.Module{}).
			Where("id IN ? AND deleted_at IS NULL", moduleIDs).
			Count(&count).Error; err != nil {
			return err
		}
		if int(count) != uniqueLen(moduleIDs) {
			return gorm.ErrRecordNotFound
		}

		rows := make([]entities.TemplateModule, 0, len(moduleIDs))
		seen := map[uint]bool{}
		for _, mid := range moduleIDs {
			if seen[mid] {
				continue
			}
			seen[mid] = true
			rows = append(rows, entities.TemplateModule{
				TemplateID: templateID,
				ModuleID:   mid,
			})
		}
		if err := tx.Create(&rows).Error; err != nil {
			return err
		}
		return nil
	})
}

func (r *repository) GetTemplatePageRepository(templateID uint, pageType string) (entities.TemplatePage, error) {
	var entity entities.TemplatePage
	if err := r.db.Where("template_id = ? AND page_type = ?", templateID, pageType).
		First(&entity).Error; err != nil {
		return entities.TemplatePage{}, err
	}
	return entity, nil
}

func (r *repository) GetTemplatePagesRepository(templateID uint) ([]entities.TemplatePage, error) {
	var list []entities.TemplatePage
	if err := r.db.Where("template_id = ?", templateID).
		Order("page_type ASC").Find(&list).Error; err != nil {
		return nil, err
	}
	return list, nil
}

// UpsertTemplatePageRepository crea o actualiza la página por defecto de la
// template. La combinación (template_id, page_type) es única para registros
// vivos, así que hacemos update-or-insert manual.
func (r *repository) UpsertTemplatePageRepository(templateID uint, pageType string, puckJSON json.RawMessage) (entities.TemplatePage, error) {
	var existing entities.TemplatePage
	err := r.db.Where("template_id = ? AND page_type = ?", templateID, pageType).First(&existing).Error
	if err == nil {
		if err := r.db.Model(&existing).Updates(map[string]interface{}{
			"puck_json_default": puckJSON,
		}).Error; err != nil {
			return entities.TemplatePage{}, err
		}
		return existing, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return entities.TemplatePage{}, err
	}

	created := entities.TemplatePage{
		TemplateID:      templateID,
		PageType:        pageType,
		PuckJSONDefault: puckJSON,
	}
	if err := r.db.Create(&created).Error; err != nil {
		return entities.TemplatePage{}, err
	}
	return created, nil
}

func uniqueLen(ids []uint) int {
	seen := map[uint]bool{}
	for _, id := range ids {
		seen[id] = true
	}
	return len(seen)
}
