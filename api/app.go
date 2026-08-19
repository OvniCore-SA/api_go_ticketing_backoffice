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

// Provisión inmediata y sincrónica de SSL en el servidor
func provisionSSL(domain string) error {
	log.Printf("⚡ [SSL AUTOMÁTICO] Ejecutando Certbot y Nginx para: %s", domain)
	cmd := exec.Command("/usr/local/bin/provision-domain", domain)
	output, err := cmd.CombinedOutput()
	if err != nil {
		log.Printf("❌ Error provisionando SSL para %s: %v. Log: %s", domain, err, string(output))
		return err
	}
	log.Printf("✅ SSL e infraestructura Nginx listos para %s", domain)
	return nil
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

	api := app.Group("/api/v1/backoffice")

	// Endpoint de verificación de SSL en tiempo real que SI NO TIENE SSL, SE LO EMITE E INSTALA AL VUELO
	api.Get("/verify-ssl", func(c *fiber.Ctx) error {
		domain := c.Query("domain")
		if domain == "" {
			return c.Status(400).JSON(fiber.Map{"error": "domain query parameter is required"})
		}

		log.Printf("🔍 Verificando estado HTTPS para: %s", domain)

		// 1. Probar si el SSL ya está emitiendo HTTPS 200 OK
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false}, // Validación estricta de SSL
		}
		client := &http.Client{
			Transport: tr,
			Timeout:   2 * time.Second,
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

		// 2. SI NO TIENE SSL VÁLIDO -> EL CÓDIGO DISPARA LA PROVISIÓN DE CERTBOT E INSTALA SSL DE UNA
		log.Printf("⚠️ %s no tiene SSL activo o falló validación. Disparando emisión automática Certbot...", domain)
		if errProv := provisionSSL(domain); errProv == nil {
			// Re-probar inmediatamente tras la emisión
			respRetry, errRetry := client.Get("https://" + domain)
			if errRetry == nil && respRetry.StatusCode < 500 {
				respRetry.Body.Close()
				return c.JSON(fiber.Map{
					"domain":     domain,
					"ssl_active": true,
					"status":     respRetry.StatusCode,
					"auto_fixed": true,
				})
			}
		}

		return c.JSON(fiber.Map{
			"domain":     domain,
			"ssl_active": false,
			"error":      "DNS propagation pending or provider challenge delayed",
		})
	})

	api.Post("/tenants", func(c *fiber.Ctx) error {
		var req CreateTenantRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
		}

		log.Printf("📥 Registrando nuevo Tenant: %s (%s) con dominio: %s", req.Name, req.Slug, req.Domain)

		// DISPARO SÍNCRONO E INMEDIATO DE SSL AL CREAR EL TENANT
		if req.Domain != "" && req.DomainType == "custom" {
			provisionSSL(req.Domain)
		}

		return c.Status(201).JSON(fiber.Map{
			"status":  "created",
			"message": "Tenant registrado y provisión SSL ejecutada de una",
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
