package entities

import "gorm.io/gorm"

// User representa una cuenta interna del backoffice (SuperAdmin).
// Este servicio no maneja cuentas de Tenant Admin, Colaborador ni Comprador —
// esas viven en ticketing-core, aisladas por Producer.
type User struct {
	gorm.Model

	Name     string `gorm:"column:name;not null"`
	Email    string `gorm:"column:email;not null;uniqueIndex:idx_users_email,where:deleted_at IS NULL"`
	Password string `gorm:"column:password;not null"`
	Active   bool   `gorm:"column:active;default:true"`

	RoleID uint `gorm:"column:role_id;not null"`
	Role   Role `gorm:"foreignKey:RoleID"`
}
