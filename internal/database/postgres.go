package database

import (
	"fmt"
	"os"
	"time"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/internal/logs"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// PostgresClient contiene la instancia de base de datos PostgreSQL
type PostgresClient struct {
	*gorm.DB
}

// NewPostgresClient crea un cliente para la base de datos PostgreSQL
func NewPostgresClient() *PostgresClient {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable search_path=%s TimeZone=UTC",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SCHEMA"),
	)
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
		NowFunc: func() time.Time {
			loc, _ := time.LoadLocation("America/Argentina/Buenos_Aires")
			return time.Now().In(loc)
		},
	})
	if err != nil {
		logs.Error("cannot connect to postgres: " + err.Error())
		panic(err)
	}

	db, _ := gormDB.DB()
	db.SetMaxIdleConns(20)
	db.SetMaxOpenConns(50)
	db.SetConnMaxLifetime(time.Hour)

	return &PostgresClient{gormDB}
}
