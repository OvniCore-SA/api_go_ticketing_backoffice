package auth

import (
	"errors"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"gorm.io/gorm"
)

// AuthRepository define las operaciones de acceso a datos para las cuentas
// internas del backoffice (SuperAdmin). No hay filtro por producer_id porque
// no existe multi-tenancy dentro de este servicio.
type AuthRepository interface {
	GetUserByEmail(email string) (entities.User, error)
	GetUserByFilter(filter filters.UserFilter) (entities.User, error)
	UpdateUserDataRepository(userID uint, updateFields map[string]interface{}) error
}

type repository struct {
	db *gorm.DB
}

func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &repository{db: db}
}

func (r *repository) GetUserByEmail(email string) (entities.User, error) {
	var user entities.User
	err := r.db.
		Preload("Role").
		Where("email = ?", email).
		First(&user).Error
	if err != nil {
		return entities.User{}, err
	}
	return user, nil
}

func (r *repository) GetUserByFilter(filter filters.UserFilter) (entities.User, error) {
	query := r.db.Model(&entities.User{}).Preload("Role")

	if len(filter.SelectFields) > 0 {
		query = query.Select(filter.SelectFields)
	}
	if filter.ID > 0 {
		query = query.Where("users.id = ?", filter.ID)
	}
	if filter.Email != "" {
		query = query.Where("users.email = ?", filter.Email)
	}
	if filter.RoleCode != "" {
		query = query.
			Joins("JOIN roles ON roles.id = users.role_id AND roles.deleted_at IS NULL").
			Where("roles.code = ?", filter.RoleCode)
	}

	if filter.ID == 0 && filter.Email == "" && filter.RoleCode == "" {
		return entities.User{}, errors.New("se requiere al menos un filtro para realizar la consulta")
	}

	var user entities.User
	if err := query.First(&user).Error; err != nil {
		return entities.User{}, err
	}
	return user, nil
}

func (r *repository) UpdateUserDataRepository(id uint, updateFields map[string]interface{}) error {
	return r.db.Model(&entities.User{}).Where("id = ?", id).Updates(updateFields).Error
}
