package api

import (
	"log"
	"os"
	"os/exec"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type CreateTenantRequest struct {
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Domain      string                 `json:"domain"`
	DomainType  string                 `json:"domainType"`
	AdminEmail  string                 `json:"adminEmail"`
	Plan        string                 `json:"plan"`
	Modules     map[string]bool        `json:"modules"`
}

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

	// Handlers de Backoffice: Gestión de Tenants y Provisionador Autónomo de Dominios
	api := app.Group("/api/v1/backoffice")

	api.Post("/tenants", func(c *fiber.Ctx) error {
		var req CreateTenantRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		log.Printf("📥 Registrando nuevo Tenant: %s (%s) con dominio: %s", req.Name, req.Slug, req.Domain)

		// DISPARO AUTOMÁTICO DEL PROVISIONADOR DE DOMINIO Y SSL (Sin intervención humana)
		if req.Domain != "" && req.DomainType == "custom" {
			go func(domain string) {
				log.Printf("🚀 Ejecutando provisión autónoma de SSL para: %s", domain)
				cmd := exec.Command("/usr/local/bin/provision-domain", domain)
				output, err := cmd.CombinedOutput()
				if err != nil {
					log.Printf("⚠️ Error provisionando dominio %s: %v. Output: %s", domain, err, string(output))
				} else {
					log.Printf("✅ Dominio y SSL provisionados autónomamente para %s", domain)
				}
			}(req.Domain)
		}

		return c.Status(201).JSON(fiber.Map{
			"status":  "created",
			"message": "Tenant registrado y provisión de SSL disparada autónomamente",
			"id":      "generated-uuid",
			"domain":  req.Domain,
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8089"
	}

	log.Printf("Starting Backoffice Go API on port %s", port)
	return app.Listen(":" + port)
}
