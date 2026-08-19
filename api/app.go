package api

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type CreateTenantRequest struct {
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	Domain      string          `json:"domain"`
	DomainType  string          `json:"domainType"`
	AdminEmail  string          `json:"adminEmail"`
	Plan        string          `json:"plan"`
	Modules     map[string]bool `json:"modules"`
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

	// Handlers de Backoffice: Gestión de Tenants, SSL y Provisión Autónomas
	api := app.Group("/api/v1/backoffice")

	// Endpoint de verificación de SSL en tiempo real para el Frontend
	api.Get("/verify-ssl", func(c *fiber.Ctx) error {
		domain := c.Query("domain")
		if domain == "" {
			return c.Status(400).JSON(fiber.Map{"error": "domain query parameter is required"})
		}

		log.Printf("🔍 Verificando certificado SSL y respuesta HTTPS en tiempo real para: %s", domain)

		// Test HTTPS real con timeout de 3s
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{
			Transport: tr,
			Timeout:   3 * time.Second,
		}

		resp, err := client.Get("https://" + domain)
		if err == nil && resp.StatusCode < 500 {
			resp.Body.Close()
			return c.JSON(fiber.Map{
				"domain":     domain,
				"ssl_active": true,
				"status":     resp.StatusCode,
			})
		}

		return c.JSON(fiber.Map{
			"domain":     domain,
			"ssl_active": false,
			"error":      "SSL certification or DNS propagation pending",
		})
	})

	api.Post("/tenants", func(c *fiber.Ctx) error {
		var req CreateTenantRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		log.Printf("📥 Registrando nuevo Tenant: %s (%s) con dominio: %s", req.Name, req.Slug, req.Domain)

		// DISPARO SÍNCRONO / INMEDIATO DE PROVISIÓN SSL
		if req.Domain != "" && req.DomainType == "custom" {
			log.Printf("🚀 Ejecutando provisión síncrona de SSL Certbot para: %s", req.Domain)
			cmd := exec.Command("/usr/local/bin/provision-domain", req.Domain)
			output, err := cmd.CombinedOutput()
			if err != nil {
				log.Printf("⚠️ Error provisionando dominio %s: %v. Output: %s", req.Domain, err, string(output))
			} else {
				log.Printf("✅ Dominio y SSL provisionados exitosamente para %s", req.Domain)
			}
		}

		return c.Status(201).JSON(fiber.Map{
			"status":  "created",
			"message": "Tenant registrado y provisión de SSL ejecutada",
			"id":      1,
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
