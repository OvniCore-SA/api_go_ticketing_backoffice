package database

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
)

// AutoMigrateAll aplica AutoMigrate a todas las entidades del backoffice.
// Solo debe llamarse cuando este servicio maneja su propio esquema (dev/local).
// En producción, el esquema es owned por ticketing-shared.
func (c *PostgresClient) AutoMigrateAll() error {
	return c.AutoMigrate(
		&entities.Role{},
		&entities.User{},
		&entities.Plan{},
		&entities.Module{},
		&entities.ComponentVariant{},
		&entities.Template{},
		&entities.TemplateModule{},
		&entities.TemplatePage{},
		&entities.Producer{},
		&entities.ProducerModule{},
		&entities.ProducerComponentVariant{},
		&entities.DesignTokens{},
		&entities.PageTemplate{},
		&entities.Domain{},
		&entities.Commission{},
	)
}
