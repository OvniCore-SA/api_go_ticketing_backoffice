package api

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api/handlers"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api/middlewares"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api/routes"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/internal/database"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/auth"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/commissions"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/design_tokens"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/domains"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/modules"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/page_templates"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/plans"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producer_component_variants"
	pcfg "github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producer_config"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producer_modules"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producers"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/public_storefront"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/store"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/templates"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

type CreateTenantRequest struct {
	Name       string          `json:"name"`
	Slug       string          `json:"slug"`
	Domain     string          `json:"domain"`
	DomainType string          `json:"domainType"`
	AdminEmail string          `json:"adminEmail"`
	Plan       string          `json:"plan"`
	Modules    map[string]bool `json:"modules"`
}

func provisionSSL(domain string) error {
	log.Printf("⚡ [SSL AUTOMÁTICO] Ejecutando Certbot y Nginx en el Host para: %s", domain)

	cmd := exec.Command("nsenter", "-t", "1", "-m", "-u", "-n", "-i", "/usr/local/bin/provision-domain", domain)
	output, err := cmd.CombinedOutput()
	if err != nil {
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

	// Inicializar la conexión a PostgreSQL DB si hay configuración disponible
	dbHost := os.Getenv("DB_HOST")
	if dbHost != "" {
		db := database.NewPostgresClient()

		env := os.Getenv("APP_ENV")
		if env == "development" || env == "dev" || env == "" {
			_ = db.AutoMigrateAll()
			_ = db.SeedBaseline()
		}

		// Repositorios
		producersRepo := producers.NewProducersRepository(db)
		domainsRepo := domains.NewDomainsRepository(db)
		producerModulesRepo := producer_modules.NewProducerModulesRepository(db)
		modulesRepo := modules.NewModulesRepository(db)
		variantsRepo := producer_component_variants.NewProducerComponentVariantsRepository(db)
		tokensRepo := design_tokens.NewDesignTokensRepository(db)
		pagesRepo := page_templates.NewPageTemplatesRepository(db)
		plansRepo := plans.NewPlansRepository(db)
		templatesRepo := templates.NewTemplatesRepository(db)
		commissionsRepo := commissions.NewCommissionsRepository(db)
		authRepo := auth.NewAuthRepository(db.DB)

		// Resolver de configuración efectiva de Producer
		resolver := pcfg.NewResolver(
			producersRepo,
			pcfg.NewStatusLayer(),
			pcfg.NewTemplateDefaultsLayer(templatesRepo),
			pcfg.NewPlanCapLayer(),
			pcfg.NewProducerOverridesLayer(producerModulesRepo, variantsRepo, tokensRepo, pagesRepo, modulesRepo),
		)

		// Servicios
		storeService := store.NewStoreService(producersRepo, domainsRepo, resolver)
		publicStorefrontService := public_storefront.NewPublicStorefrontService(producersRepo, domainsRepo, resolver)
		templatesService := templates.NewTemplatesService(templatesRepo)
		producersService := producers.NewProducersService(producersRepo, plansRepo, templatesService)
		producerModulesService := producer_modules.NewProducerModulesService(producerModulesRepo, modulesRepo, producersService)
		producerVariantsService := producer_component_variants.NewProducerComponentVariantsService(variantsRepo, modulesRepo, producersService)
		designTokensService := design_tokens.NewDesignTokensService(tokensRepo, producersService)
		pageTemplatesService := page_templates.NewPageTemplatesService(pagesRepo, producersService)
		domainsService := domains.NewDomainsService(domainsRepo, producersService)
		commissionsService := commissions.NewCommissionsService(commissionsRepo, producersService)
		authService := auth.NewAuthService(authRepo)
		modulesService := modules.NewModulesService(modulesRepo)
		plansService := plans.NewPlansService(plansRepo)

		// Handlers
		storeHandler := handlers.NewStoreHandler(storeService)
		publicStorefrontHandler := handlers.NewPublicStorefrontHandler(publicStorefrontService)
		producersHandler := handlers.NewProducersHandler(
			producersService,
			producerModulesService,
			producerVariantsService,
			designTokensService,
			pageTemplatesService,
			domainsService,
			commissionsService,
		)
		authHandler := handlers.NewAuthHandler(authService)
		modulesHandler := handlers.NewModulesHandler(modulesService)
		plansHandler := handlers.NewPlansHandler(plansService)
		templatesHandler := handlers.NewTemplatesHandler(templatesService)
		producerConfigHandler := handlers.NewProducerConfigHandler(resolver)

		// Middlewares
		mwManager := middlewares.NewMiddlewareManager(authService)
		authMiddleware := mwManager.ValidateToken()

		// API v1 Routes
		v1 := app.Group("/api/v1")

		// Endpoints de Store Engine (/api/v1/store/tenant-by-host)
		routes.SetupStoreRoutes(v1, storeHandler)

		// Endpoints Públicos de Storefront (/api/v1/public/...)
		routes.SetupPublicStorefrontRoutes(v1, publicStorefrontHandler)

		// Endpoints Autenticados de Backoffice
		routes.SetupAuthRoutes(v1, authHandler, authMiddleware)
		routes.SetupProducersRoutes(v1, producersHandler, authMiddleware)
		routes.SetupModulesRoutes(v1, modulesHandler, authMiddleware)
		routes.SetupPlansRoutes(v1, plansHandler, authMiddleware)
		routes.SetupTemplatesRoutes(v1, templatesHandler, authMiddleware)
		routes.SetupProducerConfigRoutes(v1, producerConfigHandler, authMiddleware)
	}

	// Legacy / Helper endpoints para backoffice SSL
	apiLegacy := app.Group("/api/v1/backoffice")

	apiLegacy.Get("/verify-ssl", func(c *fiber.Ctx) error {
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

	apiLegacy.Post("/tenants", func(c *fiber.Ctx) error {
		var req CreateTenantRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Invalid request body"})
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
