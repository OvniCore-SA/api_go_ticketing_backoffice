package entities

import "gorm.io/gorm"

// Códigos de rol soportados en este servicio. El backoffice solo opera con SuperAdmin;
// los demás roles del ecosistema (tenant_admin, colaborador, comprador) viven en
// ticketing-core.
const (
	RoleCodeSuperAdmin = "superadmin"
)

type Role struct {
	gorm.Model
	Code        string `gorm:"column:code;uniqueIndex;not null"`
	Name        string `gorm:"column:name;not null"`
	Description string `gorm:"column:description"`

	Users []User `gorm:"foreignKey:RoleID"`
}

func (Role) TableName() string { return "roles" }
