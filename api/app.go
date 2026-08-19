package api

import (
	"os"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api/handlers"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api/middlewares"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/api/routes"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/internal/database"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/internal/logs"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/auth"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/commissions"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/design_tokens"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/domains"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/modules"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/page_templates"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/plans"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producer_component_variants"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producer_config"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producer_modules"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producers"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/templates"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// SetupApp inicializa y arranca la API del backoffice.
//
// Este servicio es SOLO backoffice interno de SuperAdmin (ticketing-platform).
// No hay multi-tenancy dentro de la API: el SuperAdmin es global y opera sobre
// cualquier Producer especificando el ID en el path. No existe el header
// X-Producer-ID.
func SetupApp() *fiber.App {
	// 1. Infraestructura.
	dbClient := database.NewPostgresClient()

	// Guardarraíl crítico: AutoMigrate + SeedBaseline SOLO cuando
	// APP_ENV es explícitamente "development" o "dev". Cualquier otro
	// valor (incluido "" y "production") NO dispara migraciones — así
	// olvidarse la variable en prod no puede reformatear el esquema.
	// En prod las migraciones vienen de ticketing-shared.
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "development" || appEnv == "dev" {
		logs.Info("APP_ENV=" + appEnv + " → corriendo AutoMigrate + SeedBaseline")
		if err := dbClient.AutoMigrateAll(); err != nil {
			logs.Fatal("AutoMigrate falló: " + err.Error())
		}
		if err := dbClient.SeedBaseline(); err != nil {
			logs.Error("SeedBaseline falló: " + err.Error())
		}
	} else {
		logs.Info("APP_ENV=" + appEnv + " → esquema owned por ticketing-shared, no se ejecuta AutoMigrate")
	}

	// 2. Repositorios.
	authRepo := auth.NewAuthRepository(dbClient.DB)
	plansRepo := plans.NewPlansRepository(dbClient)
	modulesRepo := modules.NewModulesRepository(dbClient)
	templatesRepo := templates.NewTemplatesRepository(dbClient)
	producersRepo := producers.NewProducersRepository(dbClient)
	producerModulesRepo := producer_modules.NewProducerModulesRepository(dbClient)
	producerVariantsRepo := producer_component_variants.NewProducerComponentVariantsRepository(dbClient)
	tokensRepo := design_tokens.NewDesignTokensRepository(dbClient)
	pageTemplatesRepo := page_templates.NewPageTemplatesRepository(dbClient)
	domainsRepo := domains.NewDomainsRepository(dbClient)
	commissionsRepo := commissions.NewCommissionsRepository(dbClient)

	// 3. Servicios (respetando dependencias entre dominios).
	authService := auth.NewAuthService(authRepo)
	plansService := plans.NewPlansService(plansRepo)
	modulesService := modules.NewModulesService(modulesRepo)
	templatesService := templates.NewTemplatesService(templatesRepo)
	producersService := producers.NewProducersService(producersRepo, plansRepo, templatesService)
	producerModulesService := producer_modules.NewProducerModulesService(producerModulesRepo, modulesRepo, producersService)
	producerVariantsService := producer_component_variants.NewProducerComponentVariantsService(producerVariantsRepo, modulesRepo, producersService)
	tokensService := design_tokens.NewDesignTokensService(tokensRepo, producersService)
	pageTemplatesService := page_templates.NewPageTemplatesService(pageTemplatesRepo, producersService)
	domainsService := domains.NewDomainsService(domainsRepo, producersService)
	commissionsService := commissions.NewCommissionsService(commissionsRepo, producersService)

	// 3.b Pipeline del Decorator para lectura de configuración efectiva.
	// El orden importa: template defaults → overrides del producer →
	// cap del plan → semántica de status. Agregar reglas nuevas del
	// negocio es sumar una ConfigLayer más en el orden correcto — sin
	// tocar el resto del código.
	configResolver := producer_config.NewResolver(
		producersRepo,
		producer_config.NewTemplateDefaultsLayer(templatesRepo),
		producer_config.NewProducerOverridesLayer(
			producerModulesRepo,
			producerVariantsRepo,
			tokensRepo,
			pageTemplatesRepo,
			modulesRepo,
		),
		producer_config.NewPlanCapLayer(),
		producer_config.NewStatusLayer(),
	)

	// 4. Middlewares.
	mw := middlewares.NewMiddlewareManager(authService)
	authMiddleware := mw.ValidateToken()

	// 5. Handlers.
	authHandler := handlers.NewAuthHandler(authService)
	plansHandler := handlers.NewPlansHandler(plansService)
	modulesHandler := handlers.NewModulesHandler(modulesService)
	templatesHandler := handlers.NewTemplatesHandler(templatesService)
	producersHandler := handlers.NewProducersHandler(
		producersService,
		producerModulesService,
		producerVariantsService,
		tokensService,
		pageTemplatesService,
		domainsService,
		commissionsService,
	)
	producerConfigHandler := handlers.NewProducerConfigHandler(configResolver)

	// 6. Fiber app + middlewares globales.
	app := fiber.New(fiber.Config{
		BodyLimit: 20 * 1024 * 1024,
	})
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowCredentials: false,
		AllowOrigins:     os.Getenv("CORS_ALLOW_ORIGINS"),
		AllowHeaders:     "Content-Type, Authorization, Accept",
		AllowMethods:     "GET,POST,PUT,PATCH,DELETE",
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  true,
			"message": "API running",
			"env":     os.Getenv("APP_ENV"),
		})
	})
	app.Get("/", func(c *fiber.Ctx) error {
		return c.Send([]byte("API Ticketing Backoffice"))
	})

	// 7. Rutas versionadas.
	apiVersion := os.Getenv("API_VERSION")
	if apiVersion == "" {
		apiVersion = "v1"
	}
	apiv1 := app.Group("/api/" + apiVersion)

	routes.SetupAuthRoutes(apiv1, authHandler, authMiddleware)
	routes.SetupPlansRoutes(apiv1, plansHandler, authMiddleware)
	routes.SetupModulesRoutes(apiv1, modulesHandler, authMiddleware)
	routes.SetupTemplatesRoutes(apiv1, templatesHandler, authMiddleware)
	routes.SetupProducersRoutes(apiv1, producersHandler, authMiddleware)
	routes.SetupProducerConfigRoutes(apiv1, producerConfigHandler, authMiddleware)

	// 8. Servidor.
	port := os.Getenv("APP_PORT")
	if port == "" {
		port = "8888"
	}
	logs.Info("Servidor iniciado en puerto " + port)
	if err := app.Listen(":" + port); err != nil {
		logs.Fatal("Error al iniciar el servidor: " + err.Error())
	}
	return app
}
