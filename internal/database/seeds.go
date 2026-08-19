package database

import (
	"errors"
	"os"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/internal/logs"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// SeedBaseline crea (si no existen):
//  1. El rol SuperAdmin.
//  2. Un SuperAdmin inicial usando SUPERADMIN_EMAIL / SUPERADMIN_PASSWORD.
//  3. El catálogo mínimo de módulos (los declarados como ModuleCode* en entities).
//
// Todo es idempotente: correr múltiples veces no duplica ni pisa datos.
func (c *PostgresClient) SeedBaseline() error {
	if err := c.seedSuperAdminRole(); err != nil {
		return err
	}
	if err := c.seedSuperAdminUser(); err != nil {
		return err
	}
	if err := c.seedModules(); err != nil {
		return err
	}
	return nil
}

func (c *PostgresClient) seedSuperAdminRole() error {
	var role entities.Role
	err := c.Where("code = ?", entities.RoleCodeSuperAdmin).First(&role).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	role = entities.Role{
		Code:        entities.RoleCodeSuperAdmin,
		Name:        "SuperAdmin",
		Description: "Administrador global de la plataforma GoAccess",
	}
	if err := c.Create(&role).Error; err != nil {
		return err
	}
	logs.Info("seed: rol superadmin creado")
	return nil
}

func (c *PostgresClient) seedSuperAdminUser() error {
	email := os.Getenv("SUPERADMIN_EMAIL")
	password := os.Getenv("SUPERADMIN_PASSWORD")
	if email == "" || password == "" {
		logs.Info("seed: SUPERADMIN_EMAIL/PASSWORD no seteados, se omite alta inicial")
		return nil
	}

	var role entities.Role
	if err := c.Where("code = ?", entities.RoleCodeSuperAdmin).First(&role).Error; err != nil {
		return err
	}

	var user entities.User
	err := c.Where("email = ?", email).First(&user).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		return err
	}
	user = entities.User{
		Name:     "SuperAdmin",
		Email:    email,
		Password: string(hash),
		Active:   true,
		RoleID:   role.ID,
	}
	if err := c.Create(&user).Error; err != nil {
		return err
	}
	logs.Info("seed: superadmin inicial creado con email " + email)
	return nil
}

func (c *PostgresClient) seedModules() error {
	baseline := []entities.Module{
		{Code: entities.ModuleCodeTicketingCore, Name: "Ticketing Core", IsCore: true, Category: "core", Description: "Motor de venta de entradas"},
		{Code: entities.ModuleCodeCartelera, Name: "Cartelera", IsCore: true, Category: "content", Description: "Listado de eventos publicados"},
		{Code: entities.ModuleCodeGallery, Name: "Galería", Category: "content", Description: "Galería de imágenes del organizador"},
		{Code: entities.ModuleCodeFAQ, Name: "FAQ", Category: "content", Description: "Preguntas frecuentes"},
		{Code: entities.ModuleCodeMap, Name: "Mapa", Category: "content", Description: "Ubicación del evento en mapa"},
		{Code: entities.ModuleCodePromotions, Name: "Promociones", Category: "commerce", Description: "Cupones y promociones"},
		{Code: entities.ModuleCodePOS, Name: "POS", Category: "commerce", Description: "Venta presencial en puerta"},
	}

	for _, m := range baseline {
		var existing entities.Module
		err := c.Where("code = ?", m.Code).First(&existing).Error
		if err == nil {
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err := c.Create(&m).Error; err != nil {
			return err
		}
		logs.Info("seed: módulo " + m.Code + " creado")
	}
	return nil
}
