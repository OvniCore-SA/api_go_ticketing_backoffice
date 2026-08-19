# Spec de integración — Constructor de páginas / Theme Studio (GoAccess)

> Documento de resolución técnica para el módulo de diseño y configuración de páginas por cliente (`Producer`). Cubre stack, librería, modelo de datos parametrizable y los endpoints necesarios para llevarlo a desarrollo. Complementa a `CONTEXTO_PROYECTO_v2.md` y a `ticketing-shared-models.go`.

---

## 1. Alcance

Este documento resuelve **cómo se construye, guarda y sirve** la página de cada cliente: qué tecnología se usa, qué entidades la sostienen, y qué endpoints hacen falta para que el SuperAdmin arme el sitio en el backoffice y la tienda pública lo renderice.

No cubre: identidad/auth (ver v2 sección 9), checkout/pagos, ni validación de QR — esos son otros módulos con sus propios endpoints.

---

## 2. Stack y librería

| Pieza | Elección |
|---|---|
| Motor de edición visual | **Puck** — confirmar nombre de paquete vigente (`@measured/puck` o `@puckeditor/core`) en [puckeditor.com/docs](https://puckeditor.com/docs) al momento de instalar |
| Licencia | MIT — sin costo, sin vendor lock-in |
| Quién lo usa | **Solo el SuperAdmin**, en `platform.goaccess.com.ar`. El cliente/comprador nunca carga el editor, solo el resultado renderizado |
| Alternativa descartada | GrapesJS — descartada por trabajar con HTML/CSS libre en vez de componentes React precodeados (ver `CONTEXTO_PROYECTO_v2.md` sección 4 para el razonamiento completo) |

---

## 3. Modelo de datos — entidades parametrizables

Implementación completa en `ticketing-shared-models.go`. Resumen de referencia rápida:

| Entidad | Rol | Campos clave | Parametrizable vía |
|---|---|---|---|
| `Producer` | Tabla madre (cliente) | `slug`, `status`, `template_id`, `plan_id` | `ProducerStatus` (enum) |
| `Plan` | Etiqueta comercial, no funcional | `code`, `name` | — |
| `Template` | Catálogo de seeds (front+back) | `code`, `default_colors`, `default_fonts` | — |
| `TemplateModule` | Módulos por defecto de una template | `template_id`, `module_id` | — |
| `TemplatePage` | Página por defecto de una template | `template_id`, `page_type`, `puck_json_default` | `PageType` (enum) |
| `Module` | Catálogo global de funcionalidades | `code`, `is_core` | Constantes `ModuleCode*` |
| `ComponentVariant` | Disposiciones visuales por módulo | `module_id`, `code` (grid/carrusel/masonry) | — |
| `ProducerModule` | **Fuente única de verdad**: qué módulo ve el producer | `producer_id`, `module_id`, `enabled` | — |
| `ProducerComponentVariant` | Variante elegida por el producer, por módulo | `producer_id`, `module_id`, `component_variant_id` | — |
| `DesignTokens` | Colores/fuentes reales del producer (1:1) | `colors`, `fonts`, `radius`, `shadows` (json) | — |
| `PageTemplate` | Estructura Puck real y editable | `producer_id`, `page_type`, `puck_json` | `PageType` (enum) |
| `Domain` | Subdominios/dominios del producer | `domain`, `type` | `DomainType` (enum) |
| `Commission` | Histórico de comisión | `percentage`, `valid_from` | — |

Todas con soft delete (`gorm.Model`) y unicidad vía partial index (`WHERE deleted_at IS NULL`) — ver `0002_partial_indexes.sql`.

---

## 4. Endpoints — API de Plataforma (`ticketing-platform`)

Dueña de escritura de toda la configuración. Consumida exclusivamente por el frontend del SuperAdmin (`platform.goaccess.com.ar`). **Todos requieren rol `SuperAdmin`.**

### 4.1. Producers

| Método | Ruta | Descripción |
|---|---|---|
| `POST` | `/api/v1/producers` | Crea un producer. Body incluye `template_id` opcional — si viene, dispara el seed (copia módulos + página + tokens desde la template) |
| `GET` | `/api/v1/producers` | Lista, con filtro por `status` |
| `GET` | `/api/v1/producers/:id` | Detalle completo |
| `PATCH` | `/api/v1/producers/:id` | Edita `name`, `contact_email`, `status`, `plan_id` |
| `DELETE` | `/api/v1/producers/:id` | Baja lógica (`deleted_at`) |

### 4.2. Templates (catálogo)

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/api/v1/templates` | Lista templates disponibles |
| `POST` | `/api/v1/templates` | Crea una template nueva |
| `PATCH` | `/api/v1/templates/:id` | Edita nombre, colores/fuentes por defecto |
| `PUT` | `/api/v1/templates/:id/modules` | Reemplaza el set de `TemplateModule` (qué módulos trae de fábrica) |
| `PUT` | `/api/v1/templates/:id/pages/:pageType` | Define `puck_json_default` para un tipo de página |

### 4.3. Modules y variantes (catálogo global)

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/api/v1/modules` | Lista el catálogo completo de módulos |
| `GET` | `/api/v1/modules/:id/variants` | Variantes visuales disponibles para ese módulo |

*(Alta de módulos nuevos probablemente vía seed/migración, no vía UI — un módulo nuevo siempre implica código nuevo de todos modos.)*

### 4.4. Configuración real del producer

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/api/v1/producers/:id/modules` | Módulos habilitados/deshabilitados de este producer |
| `PUT` | `/api/v1/producers/:id/modules/:moduleId` | Habilita/deshabilita un módulo puntual (`{ "enabled": true }`) |
| `GET` | `/api/v1/producers/:id/component-variants` | Variantes elegidas por módulo |
| `PUT` | `/api/v1/producers/:id/component-variants/:moduleId` | Asigna una variante (`{ "component_variant_id": 3 }`) |
| `GET` | `/api/v1/producers/:id/design-tokens` | Tokens actuales |
| `PUT` | `/api/v1/producers/:id/design-tokens` | Actualiza colores/fuentes/radios/sombras |
| `GET` | `/api/v1/producers/:id/page-templates` | Lista las páginas del producer |
| `GET` | `/api/v1/producers/:id/page-templates/:pageType` | JSON de Puck de una página puntual (lo que carga el editor) |
| `PUT` | `/api/v1/producers/:id/page-templates/:pageType` | Guarda el JSON — **este es el `onPublish` de Puck** |
| `GET` \| `POST` \| `DELETE` | `/api/v1/producers/:id/domains` | CRUD de dominios/subdominios |
| `GET` \| `POST` | `/api/v1/producers/:id/commissions` | Histórico de comisión / alta de un nuevo registro |

---

## 5. Endpoints — API de Negocio (`ticketing-core`)

### 5.1. El endpoint que importa: render de la tienda pública

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/api/v1/storefront/pages/:pageType` | Devuelve el payload completo para `<Render>`: `puck_json` + `design_tokens` + lista de módulos habilitados. El `producer_id` no viaja en la URL — lo resuelve el middleware de Next.js a partir del dominio, y llega como header interno (`X-Producer-Id`) o ya resuelto server-side |

**Sin autenticación** (es la tienda pública). Este es el endpoint de mayor tráfico de todo el sistema de diseño — debe vivir detrás de caché en memoria (nivel 3, ya definido en la arquitectura general).

### 5.2. Nota importante: no hace falta un endpoint HTTP entre servicios para leer config

Como `ticketing-core` y `ticketing-platform` **comparten la misma base PostgreSQL**, `ticketing-core` puede leer directamente las tablas `producer_modules`, `design_tokens` y `page_templates` con su propia conexión GORM (usando los modelos de `ticketing-shared`), sin necesidad de una llamada HTTP a `ticketing-platform`. La regla de "dueño de escritura por tabla" solo restringe **quién escribe**, no impide que el otro servicio lea directo de la base.

El único endpoint HTTP real que necesitás en `ticketing-core` para este módulo es el de la sección 5.1, que expone esos datos ya combinados y cacheados hacia **afuera** (al navegador del comprador vía Next.js), no hacia el otro servicio.

### 5.3. El endpoint que consume el bloque Cartelera

Ya existe como parte de RF06, pero se menciona acá porque es el que el bloque `Cartelera` de Puck llama en tiempo real:

| Método | Ruta | Descripción |
|---|---|---|
| `GET` | `/api/v1/events?producer_id=&limit=` | Eventos del producer para poblar la Cartelera |

---

## 6. Ejemplos de payload

### Crear un producer con seed desde template

```http
POST /api/v1/producers
Authorization: Bearer <token-superadmin>

{
  "name": "FiestaBresh",
  "slug": "fiestabresh",
  "contact_email": "admin@fiestabresh.com",
  "template_id": 2,
  "plan_id": 1
}
```

Respuesta esperada: el `Producer` creado, con `ProducerModule`, `PageTemplate` y `DesignTokens` ya sembrados desde la template 2 (side-effect documentado, no requiere llamadas adicionales del frontend).

### Guardar una página desde el editor (onPublish de Puck)

```http
PUT /api/v1/producers/14/page-templates/home
Authorization: Bearer <token-superadmin>

{
  "puck_json": { "content": [ /* estructura completa de Puck */ ], "root": {} }
}
```

### Respuesta del endpoint de storefront (lo que consume `<Render>`)

```http
GET /api/v1/storefront/pages/home
```
```json
{
  "puck_json": { "content": [ /* ... */ ], "root": {} },
  "design_tokens": { "colors": { "accent": "#D85A30" }, "fonts": { "body": "Georgia, serif" } },
  "enabled_modules": ["ticketing_core", "cartelera", "promociones", "pos"]
}
```

---

## 7. Autenticación por grupo de endpoints

| Grupo | Auth requerida |
|---|---|
| Todo `/api/v1/producers*`, `/api/v1/templates*`, `/api/v1/modules*` en `ticketing-platform` | Rol `SuperAdmin` (ver RF13) |
| `/api/v1/storefront/*` en `ticketing-core` | Ninguna — público |
| `/api/v1/events` (consumido por Cartelera) | Ninguna — público, ya expone solo datos publicables |

---

## 8. Checklist antes de pasar a desarrollo

- [ ] Confirmar nombre de paquete de Puck vigente (`@measured/puck` vs `@puckeditor/core`)
- [ ] Correr `0002_partial_indexes.sql` después del primer `AutoMigrate`
- [ ] Definir el mecanismo exacto de propagación de `X-Producer-Id` desde el middleware de Next.js hacia `ticketing-core` (header interno vs. resolución server-side)
- [ ] Decidir si `PUT /producers/:id/page-templates/:pageType` valida el `puck_json` contra el `config.components` registrado en el backend, para evitar guardar referencias a bloques que ya no existen
- [ ] Definir política de caché/invalidación del endpoint `/api/v1/storefront/pages/:pageType` cuando el SuperAdmin publica un cambio (ver RNF14, consistencia de caché en 3 niveles)
