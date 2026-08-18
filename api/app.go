package api

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func StartApp() error {
	app := fiber.New(fiber.Config{
		AppName: "Ticketing SaaS Platform - Backoffice & Management API",
	})

	app.Use(logger.New())
	app.Use(cors.New())

	// Healthcheck
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"service": "api_go_ticketing_backoffice",
		})
	})

	// Handlers de Backoffice: Gestión de Tenants y Subscripciones
	api := app.Group("/api/v1/backoffice")

	api.Get("/tenants", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "success",
			"tenants": []fiber.Map{
				{"id": 1, "name": "GymAccess Ticketing", "slug": "gymaccess", "plan": "pro"},
				{"id": 2, "name": "Lollapalooza Argentina", "slug": "lollapalooza", "plan": "enterprise"},
			},
		})
	})

	api.Post("/tenants", func(c *fiber.Ctx) error {
		return c.Status(201).JSON(fiber.Map{
			"status":  "created",
			"message": "Tenant registrado exitosamente desde el Backoffice",
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8089"
	}

	log.Printf("Starting Backoffice Go API on port %s", port)
	return app.Listen(":" + port)
}
