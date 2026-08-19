# Arquitectura Go API — TelCo Clean Architecture

## Visión General

Esta API sigue los principios de **Clean Architecture** con un enfoque **Domain-Driven Design (DDD)**, separando claramente las responsabilidades en capas bien definidas. No utiliza frameworks de inyección de dependencias — toda la composición es manual via constructor injection en `api/app.go`.

**Stack tecnológico:**
- Framework HTTP: [Fiber v2](https://gofiber.io/)
- ORM: [GORM](https://gorm.io/) + Postgress
- HTTP Client: [Resty v2](https://github.com/go-resty/resty)
- Go 1.23+

---

## Estructura de Carpetas

```
{project}/
├── api/                          # Capa HTTP (entrada/salida)
│   ├── app.go                    # Bootstrap: DI container, middleware, rutas
│   ├── handlers/                 # Procesamiento de requests HTTP
│   │   └── {domain}_handler.go
│   ├── middlewares/              # Middleware HTTP
│   │   ├── auth.go               # Autenticación por API Key
│   └── routes/                   # Definición de rutas por dominio
│       └── {domain}.go
│
├── internal/                     # Paquetes internos (no exportados)
│   ├── config/
│   │   └── env.go                # Variables de entorno con defaults
│   ├── database/
│   │   └── postgres.go              # Setup cliente postgres/GORM
│   ├── logs/
│   │   └── logs.go               # Utilidad de logging

│
├── pkg/                          # Paquetes públicos reutilizables
│   ├── commons/
│   │   └── tools.go              # Utilidades comunes (UUID, imágenes, strings)
│   ├── domains/                  # Lógica de negocio por dominio
│   │   └── {domain}/
│   │       ├── {domain}_service.go      # Interfaz + implementación del servicio
│   │       └── {domain}_repository.go  # Interfaz + implementación del repositorio
│   ├── dtos/                     # Data Transfer Objects
│   │   ├── {domain}dtos/
│   │   │   ├── request.go        # DTOs de entrada con validaciones
│   │   │   ├── response.go       # DTOs de salida
│   │   │   └── errors.go         # Errores de dominio tipados
│   │   └── utils/
│   │       └── constants/
│   │           └── states.go     # Estados y outcomes del proceso
│   ├── entities/                 # Modelos de base de datos (GORM models)
│   │   └── {domain}.go
│   └── filtros/                  # Filtros y helpers de queries
│       └── utils/
│           └── paginacion.go
│
├── go.mod
├── go.sum
└── main.go                       # Entry point: llama a api.SetupApp()
```

---

## Capas de la Arquitectura

```
┌─────────────────────────────────────────────────┐
│  API Layer  (api/)                              │
│  Handlers · Routes · Middlewares                │
│  → Solo HTTP: parsea request, llama service,    │
│    formatea response                            │
└────────────────────┬────────────────────────────┘
                     │ usa interfaces de
┌────────────────────▼────────────────────────────┐
│  Service Layer  (pkg/domains/*/service)         │
│  → Lógica de negocio, orquestación             │
│  → No conoce HTTP ni DB directamente            │
└────────────────────┬────────────────────────────┘
                     │ usa interfaces de
┌────────────────────▼────────────────────────────┐
│  Repository Layer  (pkg/domains/*/repository)   │
│  → Acceso a datos: DB local o APIs externas     │
│  → Dos implementaciones posibles:               │
│     - Local: GORM/Postgress                         │
│     - Remote: HTTP clients                      │
└────────────────────┬────────────────────────────┘
                     │ usa
┌────────────────────▼────────────────────────────┐
│  Entity Layer  (pkg/entities/)                  │
│  → Modelos GORM (tablas de DB)                 │
│  → Soft deletes via gorm.Model                  │
└─────────────────────────────────────────────────┘
```

### Regla de dependencias
- Las capas superiores dependen de las inferiores **únicamente via interfaces**.
- Las capas inferiores **nunca** conocen a las superiores.
- Los DTOs fluyen entre capas: `Request DTO → Service → Response DTO`.

---

## Inyección de Dependencias

**Manual constructor injection** en `api/app.go` → función `SetupApp()`.

Orden de composición:

```go
// 1. Infraestructura
db := database.NewPostgressClient()

// 2. Repositorios de datos
domainRepo := domain.NewDomainRepository(db)

// 3. Servicios (componen repositorios)
domainService := domain.NewDomainService(domainRepo)

// 4. Middleware
authMiddleware := middlewares.NewAuthMiddleware(db)

// 5. Rutas (componen handlers con servicios)
routes.DomainRoutes(v1, authMiddleware, domainService)
```

---

## Patrones de Código

### Interfaz de Servicio

```go
// pkg/domains/{domain}/{domain}_service.go

type Service interface {
    CreateSomethingService(ctx context.Context, req dtos.RequestCreate) (dtos.ResponseCreate, error)
    GetSomethingService(ctx context.Context, id uint) (dtos.ResponseGet, error)
}

type service struct {
    repo Repository
}

func NewDomainService(repo Repository) Service {
    return &service{repo: repo}
}

func (s *service) CreateSomethingService(ctx context.Context, req dtos.RequestCreate) (dtos.ResponseCreate, error) {
    if err := req.Validate(); err != nil {
        return dtos.ResponseCreate{}, err
    }
    // lógica de negocio
    entity, err := s.repo.CreateSomethingRepository(...)
    // mapear a DTO de respuesta
    return dtos.ResponseCreate{...}, nil
}
```

### Interfaz de Repositorio

```go
// pkg/domains/{domain}/{domain}_repository.go

type Repository interface {
    CreateSomethingRepository(entity entities.Something) (entities.Something, error)
    GetSomethingByIDRepository(id uint) (entities.Something, error)
}

type repository struct {
    db *database.PostgressClient
}

func NewDomainRepository(db *database.PostgressClient) Repository {
    return &repository{db: db}
}

func (r *repository) CreateSomethingRepository(entity entities.Something) (entities.Something, error) {
    result := r.db.DB.Create(&entity)
    return entity, result.Error
}
```

### Handler

```go
// api/handlers/{domain}_handler.go

type DomainHandler struct {
    service domain.Service
}

func NewDomainHandler(service domain.Service) *DomainHandler {
    return &DomainHandler{service: service}
}

func (h *DomainHandler) CreateSomething(c *fiber.Ctx) error {
    var req dtos.RequestCreate
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  false,
            "message": "payload inválido",
            "error":   "INVALID_PAYLOAD",
        })
    }

    result, err := h.service.CreateSomethingService(c.Context(), req)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
            "status":  false,
            "message": err.Error(),
            "error":   "DOMAIN_ERROR",
        })
    }

    return c.Status(fiber.StatusCreated).JSON(fiber.Map{
        "status": true,
        "data":   result,
    })
}
```

### DTO con validación

```go
// pkg/dtos/{domain}dtos/request.go

type RequestCreate struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (r RequestCreate) Validate() error {
    if commons.StringIsEmpty(r.Name) {
        return ErrNameRequired
    }
    return nil
}
```

```go
// pkg/dtos/{domain}dtos/errors.go

var (
    ErrNameRequired  = errors.New("el nombre es requerido")
    ErrEmailRequired = errors.New("el email es requerido")
)
```

### Entidad GORM

```go
// pkg/entities/{domain}.go

type Something struct {
    gorm.Model
    Name   string `gorm:"column:name;not null"`
    Status string `gorm:"column:status;default:PENDING"`
}
```

### Formato de Respuesta Estándar

```json
// Éxito
{
    "status": true,
    "data": { ... }
}

// Error
{
    "status": false,
    "message": "Mensaje legible para el usuario",
    "error": "ERROR_CODE_SNAKE_UPPER"
}
```

---

## Configuración de Entorno

```go
// internal/config/env.go

var (
    DBHost     = getEnv("DB_HOST", "localhost")
    DBPort     = getEnv("DB_PORT", "3306")
    DBUser     = getEnv("DB_USER", "root")
    DBPassword = getEnv("DB_PASSW", "")
    DBName     = getEnv("DB_NAME", "my_db")

    CORS_ALLOW_ORIGINS = getEnv("CORS_ALLOW_ORIGINS", "http://localhost:3000")
)

func getEnv(key, fallback string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return fallback
}
```

---

## Middleware

### Cadena de middleware (en app.go)

```
Request
  ↓ logger.New()           # Log HTTP request/response
  ↓ recover.New()          # Panic recovery
  ↓ cors.New()             # CORS headers
  ↓ TrackingMiddleware     # UUID v7 correlación + audit log
  ↓ Route /api/v1
      ↓ AuthMiddleware     # API Key validation
      ↓ Handler
```

### Auth Middleware

```go
func (m *AuthMiddleware) ValidateAPIKey(c *fiber.Ctx) error {
    apiKey := c.Get("Authorization")
    // validar contra DB
    if !valid {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "status":  false,
            "message": "API Key inválida",
            "error":   "UNAUTHORIZED",
        })
    }
    c.Locals("api_client", client)
    return c.Next()
}
```

---

## Operaciones Asíncronas

Para operaciones de auditoría o logging que no deben bloquear el flujo principal, se usa goroutines:

```go
go func() {
    if err := s.trackingService.RecordRequestService(trackingData); err != nil {
        logs.Error("error registrando tracking: " + err.Error())
    }
}()
```

---

## Convenciones de Nomenclatura

| Elemento | Convención | Ejemplo |
|---------|-----------|---------|
| Archivos | snake_case | `persona_handler.go` |
| Paquetes | lowercase singular | `package persona` |
| Interfaces | PascalCase sin prefijo | `type Service interface` |
| Structs (privados) | camelCase | `type service struct` |
| Constructores | `New{Type}` | `NewPersonaService(...)` |
| Métodos de servicio | `{Action}{Domain}Service` | `CreateIdentificationService` |
| Métodos de repo | `{Action}{Domain}Repository` | `GetPersonaByDNIRepository` |
| DTOs Request | `Request{Action}` | `RequestCreateIdentification` |
| DTOs Response | `Response{Action}` | `ResponseCreateIdentification` |
| Constantes de error | `Err{Description}` | `ErrDNIRequired` |

---

## Prompt Base para Nueva API

Usá el siguiente prompt para inicializar una nueva API con esta misma arquitectura:

---

```
Crea una API REST en Go usando la siguiente arquitectura Clean Architecture con DDD.

## Stack
- Framework HTTP: Fiber v2 (github.com/gofiber/fiber/v2)
- ORM: GORM (gorm.io/gorm) con driver Postgress (gorm.io/driver/Postgress)
- HTTP Client: Resty v2 (github.com/go-resty/resty/v2) [solo si hay integraciones externas]
- UUID: github.com/google/uuid
- Go 1.23+

## Estructura de carpetas a generar

{project_name}/
├── main.go
├── go.mod
├── api/
│   ├── app.go
│   ├── handlers/
│   │   └── {domain}_handler.go
│   ├── middlewares/
│   │   ├── auth.go
│   │   ├── tracking.go
│   │   └── logger.go
│   └── routes/
│       └── {domain}.go
├── internal/
│   ├── config/
│   │   └── env.go
│   ├── database/
│   │   └── Postgress.go
│   └── logs/
│       └── logs.go
└── pkg/
    ├── commons/
    │   └── tools.go
    ├── domains/
    │   └── {domain}/
    │       ├── {domain}_service.go
    │       └── {domain}_repository.go
    ├── dtos/
    │   └── {domain}dtos/
    │       ├── request.go
    │       ├── response.go
    │       └── errors.go
    └── entities/
        └── {domain}.go

## Reglas de arquitectura

1. Separación estricta por capas: API → Service → Repository → Entity
2. Cada capa se comunica con la siguiente SOLO a través de interfaces
3. Inyección de dependencias manual via constructores en api/app.go
4. NO usar frameworks de DI (wire, fx, dig)
5. DTOs separados por dominio con método Validate() en los request DTOs
6. Errores de dominio como variables exportadas (var ErrXxx = errors.New(...))
7. Formato de respuesta estándar:
   - Éxito: {"status": true, "data": {...}}
   - Error: {"status": false, "message": "...", "error": "ERROR_CODE"}
8. Configuración via variables de entorno con defaults en internal/config/env.go
9. Pool de conexiones Postgress: MaxOpenConns=50, MaxIdleConns=20, ConnMaxLifetime=1h
10. CORS configurado desde variable de entorno CORS_ALLOW_ORIGINS
11. Middleware de autenticación por API Key en header Authorization
12. Tracking/auditoría de requests con UUID v7 (time-ordered)
13. Operaciones de audit log asíncronas con goroutines para no bloquear respuesta
14. Soft deletes usando gorm.Model en todas las entidades

## Convenciones de nomenclatura
- Archivos: snake_case
- Paquetes: lowercase singular (package persona, package tracking)
- Interfaces: PascalCase (Service, Repository)
- Structs internos: camelCase (service, repository)
- Constructores: New{Type} → NewPersonaService
- Métodos servicio: {Action}{Domain}Service → CreateIdentificationService
- Métodos repo: {Action}{Domain}Repository → GetPersonaByIDRepository
- DTOs: Request{Action} / Response{Action}
- Errores: ErrDescripcion → ErrDNIRequired

## Dominio inicial a implementar

Nombre del dominio: {domain_name}
Entidad principal: {entity_fields}
Endpoints:
  - POST   /api/v1/{domain}         → Create
  - GET    /api/v1/{domain}/:id     → GetByID
  - PUT    /api/v1/{domain}/:id     → Update
  - DELETE /api/v1/{domain}/:id     → Delete (soft)
  - GET    /api/v1/{domain}         → List (con paginación)

Generá todo el código siguiendo exactamente estos patrones, sin agregar abstracciones extra
ni dependencias adicionales que no sean necesarias para el dominio solicitado.
```

---

## Checklist al agregar un nuevo dominio

- [ ] Crear `pkg/entities/{domain}.go` con el modelo GORM
- [ ] Crear `pkg/dtos/{domain}dtos/request.go` con validaciones
- [ ] Crear `pkg/dtos/{domain}dtos/response.go`
- [ ] Crear `pkg/dtos/{domain}dtos/errors.go` con errores tipados
- [ ] Crear `pkg/domains/{domain}/{domain}_repository.go` (interfaz + impl)
- [ ] Crear `pkg/domains/{domain}/{domain}_service.go` (interfaz + impl)
- [ ] Crear `api/handlers/{domain}_handler.go`
- [ ] Crear `api/routes/{domain}.go`
- [ ] Registrar en `api/app.go`: repo → service → middleware → routes
- [ ] Agregar migración de tabla en `internal/database/Postgress.go`
