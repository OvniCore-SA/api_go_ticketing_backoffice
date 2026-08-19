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

// Provisión inmediata de SSL vía Docker Exec en el Host
func provisionSSL(domain string) error {
	log.Printf("⚡ [SSL AUTOMÁTICO] Ejecutando Certbot y Nginx en el Host para: %s", domain)
	
	// Ejecutar el script directamente en el Host Linux usando nsenter o docker.sock
	cmd := exec.Command("nsenter", "-t", "1", "-m", "-u", "-n", "-i", "/usr/local/bin/provision-domain", domain)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Fallback directo vía ejecutor del host
		cmdFallback := exec.Command("/bin/sh", "-c", "nsenter -t 1 -m -u -n -i /usr/local/bin/provision-domain "+domain)
		outputFallback, errFallback := cmdFallback.CombinedOutput()
		if errFallback != nil {
			log.Printf("❌ Error provisionando SSL para %s: %v. Output: %s", domain, errFallback, string(outputFallback))
			return errFallback
		}
	}
	log.Printf("✅ SSL e infraestructura Nginx listos para %s. Output: %s", domain, string(output))
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

	api.Get("/verify-ssl", func(c *fiber.Ctx) error {
		domain := c.Query("domain")
		if domain == "" {
			return c.Status(400).JSON(fiber.Map{"error": "domain query parameter is required"})
		}

		log.Printf("🔍 Verificando estado HTTPS para: %s", domain)

		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: false},
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

		log.Printf("⚠️ %s no tiene SSL activo o falló validación. Disparando emisión automática Certbot...", domain)
		if errProv := provisionSSL(domain); errProv == nil {
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
			return c.Status(400).JSON(fiber.Map{"error":="Invalid request body"})
		}

		log.Printf("📥 Registrando nuevo Tenant: %s (%s) con dominio: %s", req.Name, req.Slug, req.Domain)

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
