# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

**API Backoffice de GoAccess (`ticketing-platform`)** — Go + Fiber + GORM + PostgreSQL.

Único responsable de administrar la configuración de la plataforma multi-tenant: `Producer` (tenant), planes, dominios, comisiones, y el motor de diseño (`Template`, `Module`, `ComponentVariant`, `DesignTokens`, `PageTemplate` — la parte del Theme Studio que consume Puck en el frontend).

**Este servicio es interno.** El único rol con acceso es `SuperAdmin`. El ticketing (eventos, órdenes, tickets, QR) y las cuentas de Tenant Admin / Colaborador / Comprador viven en el otro servicio (`ticketing-core`), no acá.

Documentación de contexto (tratar como fuente de verdad cuando exista):
- `CONTEXTO_PROYECTO.md` — visión de producto, modelo multi-tenant, arquitectura de los dos servicios Go, regla de "dueño de escritura por tabla".
- `spec-theme-studio-integracion.md` — spec del constructor de páginas y del catálogo de módulos/variantes/templates.
- `MI_ESTILO_GO_BACKEND.md` — patrón de código que sigue este repo verbatim; §15 es el checklist para agregar un dominio.
- `ARQUITECTURA_GO_API.md` — referencia genérica de arquitectura.

## Commands

```bash
# Correr local (carga .env por godotenv; APP_PORT default 8888).
# En APP_ENV=development/dev/"" se ejecutan AutoMigrate + SeedBaseline.
go run .

# Build.
go build -o server .

# Docker (Dockerfile pin: golang:1.22-alpine; go.mod pide 1.25.7,
# actualizar el Dockerfile antes de usarlo en prod).
docker build -t ticketing-backoffice .

# Dependencies.
go mod tidy
```

No hay tests todavía. `.github/workflows/deploy.yml.disabled` — CI/CD está apagado a propósito.

### Required environment variables

- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SCHEMA` — DSN de Postgres + `search_path`.
- `JWT_SECRET_KEY` — clave HS256 para access + refresh tokens.
- `CORS_ALLOW_ORIGINS` — origins permitidos separados por coma.
- `API_VERSION` (default `v1`), `APP_PORT` (default `8888`), `APP_ENV`.
- `SUPERADMIN_EMAIL`, `SUPERADMIN_PASSWORD` — para el seed del SuperAdmin inicial.

## Architecture

### Ciclo de request `/api/{version}/*`

```
Fiber → logger / recover / cors
     → apiv1 group
         → Setup*Routes(apiv1, handler, authMiddleware)
             → (público) POST /auth/login
             → (protegido, requiere JWT SuperAdmin)
                 ValidateToken() → handler → service → repository → gorm
```

**No hay middleware `ResolveProducer` ni header `X-Producer-ID`.** Ese esquema es del otro servicio (`ticketing-core`), donde la tienda pública resuelve el tenant por dominio. Acá el SuperAdmin es global y el `producer_id` viaja en el path (`/producers/:id/...`).

### `ValidateToken` (`api/middlewares/auth.go`)

Garantías que impone (preservar):
1. Solo firma HMAC — rechaza `alg:none` y variantes RSA.
2. Recarga el usuario desde DB filtrando por `roles.code = superadmin`. Si dejó de ser SuperAdmin o fue desactivado, el token queda inválido inmediatamente.
3. El usuario autenticado queda en `c.Locals("user")` como `authdtos.ResponseUser`. Los handlers lo leen con `authUserFromCtx(c)`.

Los tokens (48h access, 5d refresh) se emiten en `pkg/domains/auth/auth_service.go#GetTokensService`.

### Wiring (`api/app.go`)

No hay DI container. `SetupApp()` compone todo en orden fijo: DB → (dev only: AutoMigrate + SeedBaseline) → repos → services → middleware → handlers → app Fiber → routes → `Listen`. Cuando agregues un dominio nuevo, insertá su repo/service/handler en ese orden y su `routes.SetupXRoutes` dentro de `apiv1`.

### Capas (resumido — ver `MI_ESTILO_GO_BACKEND.md`)

- `pkg/entities/` — modelos GORM. `gorm.Model` embebido (soft delete). FK nullable como `*uint`. Constantes de estado arriba del struct. Unicidad multi-registro con partial index `WHERE deleted_at IS NULL` para permitir re-crear tras baja lógica.
- `pkg/filters/{domain}_filter.go` — struct de filtros para queries del repo.
- `pkg/dtos/{domain}dtos/` — `Request*` con `Validate()`/`ToEntity()`; `Response*` con `FromEntity()`/`FromEntities()`. `errors.go` con sentinels de dominio.
- `pkg/domains/{domain}/` — `Repository` interface + `repository` privado + `NewXRepository`, y el `Service` paralelo. Los updates usan `map[string]interface{}` (nunca `Save(&entity)` para no pisar columnas con zero-values).
- `api/handlers/` — parsean input (`BodyParser`/`QueryParser`/`parseUintParam`), llaman al service, envuelven la respuesta. Los errores se pasan por `handleServiceError` (dominios genéricos) o `handleAuthError` (dominio auth con `AuthError` tipado). Envelope estándar: `{"status": true, "data": ...}` / `{"status": false, "message": ...}`.
- `api/routes/` — un `SetupXRoutes(router, handler, authMiddleware)` por dominio.

El dominio `auth` sigue una convención propia (`IAuthService`, `AuthError{Code, Message}` traducido por `handleAuthError`). Los dominios de negocio siguen el naming plano (`Service`, `Repository`) del estilo general.

### Motor de diseño (Theme Studio)

Modelo que sostiene la personalización por Producer, todo owned por este servicio:

- **Catálogos globales:** `Module` (código de bloque), `ComponentVariant` (disposición visual del módulo), `Template` (preset colores + fuentes + set de módulos + páginas Puck).
- **Preset a Producer (seed):** `producers.SeedProducerFromTemplateRepository` copia en una tx: `TemplateModule → ProducerModule` (enabled=true), `TemplatePage → PageTemplate`, y los `default_*` de la Template → `DesignTokens`. Si el Producer se crea sin `template_id`, se crea un `DesignTokens` vacío para mantener 1:1.
- **Configuración real:** `ProducerModule` es la fuente única de verdad de qué módulos ve un tenant. `ProducerComponentVariant` guarda qué variante eligió por módulo. `DesignTokens` y `PageTemplate` (1 por page_type) son los datos que consume la tienda pública desde `ticketing-core`.
- **Módulos core:** los `Module.IsCore = true` no pueden deshabilitarse en `ProducerModule` (`producer_modules_service` rechaza el toggle con 409).

### Comisiones

`Commission` guarda el histórico. Cada alta cierra la comisión vigente del producer (setea su `ValidTo`) y crea el nuevo registro con `ValidTo NULL` — todo en la misma transacción (`CreateWithCloseRepository`).

### Base de datos

`internal/database/postgres.go` devuelve `*PostgresClient` (embebe `*gorm.DB`). Timezone `America/Argentina/Buenos_Aires` para `NowFunc`. Pool: 20 idle / 50 open / 1h max lifetime.

En prod, el esquema es owned por `ticketing-shared`; este servicio no corre migraciones. `AutoMigrateAll` + `SeedBaseline` solo se ejecutan en `APP_ENV=development/dev/""`.

`SeedBaseline` es idempotente:
- rol `superadmin`;
- SuperAdmin inicial (si `SUPERADMIN_EMAIL` + `SUPERADMIN_PASSWORD` están seteados);
- catálogo mínimo de módulos (`ModuleCode*` en `pkg/entities/module.go`).

### Copy al usuario

Español (Argentina), tono cercano, siempre proponer un próximo paso. Mensajes de auth centralizados en `pkg/domains/auth/auth.messages.go` y los del middleware en `api/middlewares/messages.go`.
