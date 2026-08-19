# GoAccess — Plataforma de Ticketing y Gestión de Eventos Multi-Tenant White-Label

> Documento de contexto para desarrollo. Última actualización: en construcción activa.
> Este archivo resume todas las decisiones de arquitectura, requerimientos y convenciones acordadas hasta el momento. Es la fuente de verdad para dar contexto a herramientas de desarrollo asistido por IA (Claude Code, Cursor, etc.).

---

## 1. Visión general

Plataforma SaaS de ticketing y gestión de eventos **multi-tenant**, inspirada en el modelo de Fanz (Argentina). Permite que productoras de eventos y organizadores (**tenants**) operen bajo su propia marca, manteniendo control de sus datos de audiencia, catálogo de eventos y transacciones financieras.

La experiencia es **totalmente white-label**: cada organizador usa su propio subdominio o un dominio personalizado independiente. El comprador final (fan) nunca percibe la plataforma subyacente — solo ve la marca del organizador.

**Nombre del producto:** GoAccess
**Dominio raíz:** `goaccess.com.ar`

**Referencias de mercado (competencia):**
- Modelo igual al nuestro (white-label, dominio propio, split payment directo): **Fanz**, **Enigmatickets**.
- Modelo marketplace (catálogo centralizado, NO es nuestro modelo): Ticketek, EntradaUno, TicketHoy, Livepass.

---

## 2. Glosario

| Término | Significado |
|---|---|
| Tenant | Productora/organizador que opera su propia tienda dentro de la plataforma |
| White-label | Marca blanca: el tenant presenta la plataforma como propia, con su dominio y estética |
| SuperAdmin | Equipo interno de GoAccess; administra tenants, planes, diseño y comisiones |
| Tenant Admin | Administrador de una productora; gestiona sus eventos, entradas, cupones, compradores |
| Colaborador | Staff de puerta que valida entradas escaneando QR |
| Comprador / Fan | Usuario final que compra entradas en la tienda de un tenant |
| BFF | Backend for Frontend: capa Next.js que sirve la interfaz y orquesta llamadas al backend |
| Hot sale | Pico de venta masiva con alta concurrencia sobre un mismo evento |

---

## 3. Stack tecnológico

| Capa | Tecnología |
|---|---|
| Frontend / BFF | **Next.js 15** (App Router), sobre Vercel Edge, basado en [Vercel Platforms Starter Kit](https://vercel.com/templates/next.js/platforms-starter-kit) |
| Motor visual | Puck (`@measured/puck`) — editor drag-and-drop + renderizado SSR basado en JSON |
| Estilos/variantes | Class Variance Authority (CVA) sobre Tailwind CSS; Design Tokens como variables CSS dinámicas |
| Backend | Golang + Fiber, GORM como ORM, contenedores Docker con autoescalado |
| Base de datos | PostgreSQL administrado (Neon / Supabase / AWS RDS) + PgBouncer |
| Caché / mensajería | Redis (Upstash / Vercel KV) |
| Almacenamiento | **AWS S3** (logos, banners, galería, tickets PDF) |
| Pasarela de pagos | **Mercado Pago** (modelo marketplace/Connect, split payments) |

### Por qué Next.js y no un React SPA común (decisión justificada)

Next.js es obligatorio, no una preferencia, porque:
1. El middleware de resolución de tenant necesita correr en el servidor/Edge **antes** de que exista cualquier HTML (intercepta el header `Host`).
2. RNF02 exige resolución de dominio <5ms — solo alcanzable corriendo en el Edge, no en un bundle que se ejecuta en el navegador.
3. La tienda pública necesita SSR/ISR para SEO y velocidad de carga (crítico en conversión de venta de entradas). Una SPA no tiene renderizado en servidor.

Un React común solo sería válido para un backoffice aislado sin necesidad de SEO ni resolución de dominios — pero en este proyecto el backoffice vive en el mismo proyecto Next.js que la tienda pública.

---

## 4. Arquitectura general

### 4.1. Vista de capas

```
┌─────────────────────────────────────────────────────────┐
│  FRONTEND / BFF — Un solo proyecto Next.js 15            │
│  El middleware resuelve tenant según el dominio entrante  │
│                                                             │
│  platform.goaccess.com.ar     → SuperAdmin                │
│  fiestabresh.com               → Tienda pública            │
│  fiestabresh.com/admin         → Tenant Admin               │
│  (ruta o subdominio)           → App de escáner (PWA)       │
└──────────────┬──────────────────────────┬─────────────────┘
               │                          │
   ┌───────────▼───────────┐   ┌──────────▼─────────────┐
   │  ticketing-platform    │   │   ticketing-core        │
   │  (Go, backoffice)      │   │   (Go + Fiber, negocio) │
   │  api.backoffice.       │   │   api.ticketing.        │
   │  goaccess.com.ar       │   │   goaccess.com.ar       │
   │                        │   │                         │
   │  ESCRIBE:              │   │  ESCRIBE:               │
   │  - tenants             │   │  - eventos, inventario  │
   │  - planes              │   │  - órdenes, tickets     │
   │  - diseño (tokens/Puck)│   │  - pagos (Mercado Pago) │
   │  - comisiones          │   │  - CRM compradores      │
   │                        │   │  - validaciones QR      │
   │  LEE (métricas):       │   │  LEE (config, cacheada):│
   │  - datos transaccional.│   │  - tenants, planes,     │
   │                        │   │    diseño, comisiones   │
   └───────────┬────────────┘   └──────────┬──────────────┘
               │                          │
               └──────────┬───────────────┘
                          │
              ┌───────────▼────────────┐
              │  PostgreSQL compartido  │
              │  (dueño de escritura    │
              │   definido por tabla)   │
              └─────────────────────────┘

  ticketing-shared → paquete Go compartido con modelos y
  migraciones de entidades núcleo (fuente única del esquema)

  Redis → caché Edge (nivel 1), locks de stock, colas
  AWS S3 → PDFs, imágenes, logos, banners
  Mercado Pago → integración externa de pagos (solo desde ticketing-core)
```

### 4.2. Repositorios

| Repo | Contenido |
|---|---|
| `ticketing-core` | API de Negocio (Go + Fiber): eventos, inventario, órdenes, tickets, pagos, QR, auth de tenant/comprador |
| `ticketing-platform` | API de Plataforma (Go): backoffice del SuperAdmin — tenants, planes, diseño, comisiones |
| `ticketing-shared` | Paquete Go compartido: modelos GORM y migraciones de entidades núcleo |
| (por definir) | Repo Next.js (frontend, único, basado en Platforms Starter Kit) |

### 4.3. Dominios

| Servicio | Dominio |
|---|---|
| Frontend backoffice (SuperAdmin) | `platform.goaccess.com.ar` |
| Tienda pública del tenant | dominio propio del tenant (ej. `fiestabresh.com`) o `fiestabresh.goaccess.com.ar` |
| Panel Tenant Admin | mismo dominio del tenant + `/admin` (ej. `fiestabresh.com/admin`) — **no** usa subdominio separado |
| API de Negocio | `api.ticketing.goaccess.com.ar` |
| API de Plataforma | `api.backoffice.goaccess.com.ar` |

### 4.4. Regla de oro: dueño de escritura por tabla

Los dos servicios Go comparten una única base PostgreSQL. Para evitar acoplamiento caótico, **cada tabla tiene un único servicio que la escribe**; el otro, si la necesita, solo la lee.

- **Escribe `ticketing-platform`** (config): `tenants`, `planes`, `configuración de diseño` (tokens, Puck, feature flags), `comisiones`. → `ticketing-core` las lee y cachea agresivamente (nivel 3 de caché).
- **Escribe `ticketing-core`** (transaccional): `eventos`, `tipos de entrada/inventario`, `órdenes`, `tickets`, `pagos`, `compradores/CRM`, `validaciones QR`. → `ticketing-platform` las lee solo para métricas de facturación.

Los modelos de las entidades núcleo viven en `ticketing-shared`, importados por ambos servicios, para que nunca se dupliquen ni desincronicen.

### 4.5. Baja lógica universal

Ninguna entidad se borra físicamente. Se usa `gorm.Model` (con `DeletedAt`), aprovechando el soft delete nativo de GORM (`deleted_at IS NULL`). Aplica a tenants, eventos, usuarios y toda otra entidad.

**Dos reglas importantes:**
- **Unicidad + soft delete:** la unicidad del comprador es `(tenant_id, email)`, pero debe incluir `deleted_at` en el índice → `(tenant_id, email, deleted_at)`, para permitir re-registro tras una baja.
- **Baja lógica ≠ estado de negocio:** `deleted_at` responde "¿existe o fue eliminado?" (técnico). `status` (activo/suspendido) responde "¿cómo opera mientras existe?" (negocio). Un tenant **suspendido** no está borrado.

### 4.6. Multi-tenancy y dominios

- **Subdominios:** DNS wildcard (`*.goaccess.com.ar`) apuntando a Vercel; funcionan sin intervención manual.
- **Dominios personalizados:** al ingresar un dominio propio, el sistema invoca la API de Dominios de Vercel para aprovisionar y generar el certificado SSL.
- **Mapeo:** el middleware de Next.js lee el header `Host`, resuelve el tenant y reescribe la ruta internamente, manteniendo la URL intacta en el navegador.

### 4.7. Estrategia de caché en 3 niveles

1. **Nivel 1 — Edge (Redis):** el middleware resuelve qué tenant corresponde a cada dominio en <5ms, sin tocar la API de Go.
2. **Nivel 2 — Página (ISR de Next.js):** información visual del evento servida desde caché estática con revalidación periódica.
3. **Nivel 3 — Memoria en Go:** catálogos y datos de tenants cacheados en RAM dentro de `ticketing-core`. PostgreSQL reservado para transacciones críticas.

### 4.8. Base técnica de referencia: Vercel Platforms Starter Kit

URL: https://vercel.com/templates/next.js/platforms-starter-kit

Aporta el patrón base de resolución de subdominios (middleware → Redis `subdomain:{name}` → reescritura de ruta). **Lo que NO trae y hay que construir:** CRUD real de tenants, ruta `/admin` por tenant (el starter solo tiene `/admin` en el dominio raíz), conexión con `ticketing-core`/`ticketing-platform`, auth, RBAC, Puck, checkout.

---

## 5. Identidad, autenticación y roles

**Estado: CERRADO** (RF12, RF13)

### 5.1. Modelo de identidad del comprador

1. **Registro obligatorio para comprar** — no existe guest checkout.
2. **Usuario aislado por tenant** — unicidad `(tenant_id, email)`. El mismo email puede existir como cuenta separada en cada productora (refuerza el aislamiento white-label; cada tenant es dueño de su audiencia).
3. **Login social (Google) compatible con el aislamiento** — el `tenant_id` se resuelve desde el dominio/contexto (Edge), NO desde Google. Google solo confirma titularidad del email; la cuenta se ata al tenant del contexto. Se guarda el método de registro (password/Google) por cuenta.

### 5.2. Roles del sistema (RBAC)

| Rol | Alcance |
|---|---|
| **SuperAdmin** | Poder global de plataforma. Gestiona tenants, planes, diseño, comisiones, métricas. Vive en `ticketing-platform`, fuera del modelo de tenants. |
| **Tenant Admin** | Poder total dentro de su propio tenant: eventos, precios, cupones, CRM, gestión de su staff. Cero visibilidad fuera de su tenant. |
| **Colaborador** | Único permiso: acceder a la app de escáner. No ve tienda, panel, reportes ni datos de compradores. |
| **Comprador / Fan** | Gestiona únicamente sus propias compras y tickets. |

**Regla estructural:** cada cuenta pertenece a un único tenant `(tenant_id, email)`; no hay roles cruzados entre tenants.

### 5.3. Modelo del colaborador (escáner)

1. Login en el tenant (mismo sistema de auth, atado al dominio/contexto).
2. El sistema detecta rol `colaborador` → lo lleva a la app de escáner, no al panel ni a la tienda.
3. La app muestra los eventos asignados para escanear.
4. Al elegir un evento habilitado, se activa la cámara y se valida el QR contra `ticketing-core` (<100ms).

**Doble capa de permiso:** rol + asignación de evento. Sin evento asignado, no hay cámara habilitada. Al finalizar el evento, la asignación deja de habilitar el escaneo (sin borrar la cuenta).

---

## 6. Plantillas, modularidad y estilos

- **Feature flagging por plan:** el SuperAdmin habilita/deshabilita módulos por tenant según su plan (ej. Básico con 3 secciones vs. Premium con 5, incluyendo Galería y FAQ).
- **Design Tokens:** variables CSS globales (paleta, fuentes, bordes, sombras) guardadas como JSON asociado al tenant.
- **Variantes CVA:** cada bloque (banner, grilla, galería) admite múltiples disposiciones (grilla estándar, masonry, carrusel).
- **Editor visual (Puck):** el SuperAdmin arma la estructura arrastrando bloques; se guarda como JSON en PostgreSQL.
- **Renderizado dinámico:** la tienda recibe la lista de secciones autorizadas + tema del tenant, e itera solo sobre bloques habilitados.

---

## 7. Portales y experiencias

| Portal | Descripción |
|---|---|
| **SuperAdmin** (`platform.goaccess.com.ar`) | Interno. Gestión de tenants, dominios, Theme Studio (Puck), módulos, comisiones, métricas globales. |
| **Tenant Admin** (dominio del tenant + `/admin`) | Organizadores. Eventos, sectores, lotes, precios, stock, cupones, CRM, reportes, gestión de staff. |
| **Tienda pública white-label** (dominio del tenant) | Compradores. Cartelera, detalle de evento, galería, selección de entradas, checkout. |
| **App de escáner** (PWA) | Staff/colaboradores. Lectura de QR con validación en tiempo real. |

---

## 8. Qué debe cubrir el Frontend (Next.js)

### 8.1. Motor de ruteo
- Middleware de resolución de tenant: subdominio/dominio custom → Redis → fallback a `ticketing-platform`.
- Ruteo por rol dentro del mismo dominio del tenant (tienda vs. `/admin`), con aislamiento de sesión.
- Dominio separado para SuperAdmin con su propio login.

### 8.2. Portal SuperAdmin
- Login con 2FA.
- CRUD de tenants (RF14).
- Gestión de planes y feature flags (RF15).
- Configuración de comisiones (RF16).
- Dashboard de métricas globales (RF17).
- Theme Studio: editor Puck + Design Tokens + selector de variantes CVA.

### 8.3. Panel Tenant Admin
- Login propio del tenant.
- CRUD de eventos, sectores, lotes, precios, stock (RF06).
- Cupones y promociones (RF18).
- CRM + reportes exportables (RF11, RF20).
- Gestión de colaboradores y asignación de eventos (RF19).
- Carga de logo/banner/favicon (RF21).
- Estados del evento: borrador/publicado/finalizado (RF22).
- Onboarding de Mercado Pago (vinculación de cuenta para split payment).

### 8.4. Tienda pública
- Renderizado dinámico vía Puck (solo bloques habilitados por plan).
- Cartelera, detalle, galería (RF05).
- Registro/login del comprador (email+password y Google), aislado por tenant.
- Carrito, cupón, checkout (RF23).
- Integración de pago con Mercado Pago.
- Página "mis compras / mis tickets".

### 8.5. App de escáner
- Login del colaborador.
- Eventos asignados.
- Cámara + lectura QR contra `ticketing-core`.
- **Formato:** PWA dentro del mismo proyecto Next.js (decisión tentativa — ver sección 11).

### 8.6. Capas transversales
- Cliente de datos hacia ambas APIs (abstracción clara de a cuál service le pega cada llamada).
- Manejo de sesión/auth por rol, con `tenant_id` siempre en contexto.
- Sistema de diseño compartido (CVA), tematizable por tenant solo en la tienda pública.
- Pantallas de estado/error para escenarios de hot sale (stock agotado en checkout, pago rechazado, etc.).

### 8.7. Fuera del frontend
Toda la lógica crítica (bloqueo de stock, validación de pago, generación de QR firmado con HMAC, anti-sobreventa) vive en `ticketing-core`. El frontend dispara acciones y muestra resultados — nunca decide por sí solo sobre stock o validez de ticket.

---

## 9. Especificación de requerimientos

### 9.1. Requerimientos funcionales — base del proyecto

| ID | Nombre | Descripción |
|---|---|---|
| RF01 | Resolución multi-tenant | Servir múltiples tiendas con una única base de código vía subdominios y dominios custom |
| RF02 | Registro de dominios automático | Vincular dominios de clientes con Vercel vía su API |
| RF03 | Editor de páginas y estilos | Definir paletas, fuentes, formas y variante visual de cada módulo por cliente (Theme Studio) |
| RF04 | Habilitación modular de secciones | Activar/desactivar secciones (Hero, Cartelera, Galería, Mapa, FAQ) por tenant según plan |
| RF05 | Módulo de galería de fotos | Carga y almacenamiento en S3 con presentación configurable |
| RF06 | Gestión de eventos e inventario | Eventos con categorías de entradas, cupos por sector, fechas de venta |
| RF07 | Bloqueo temporal de stock | Reserva de entradas durante el pago (10-15 min configurable) |
| RF08 | Pagos con split | Pasarela que divide la transacción entre productora y comisión de plataforma |
| RF09 | Emisión de tickets criptográficos | PDF + QR firmado con HMAC tras confirmación de pago |
| RF10 | Validador de puerta | API de alta velocidad: validez, uso previo, pertenencia al evento |
| RF11 | CRM de compradores | Datos de contacto e historial de compra por productora |

### 9.2. Requerimientos funcionales — definidos en esta etapa

| ID | Nombre | Estado | Descripción |
|---|---|---|---|
| RF12 | Autenticación multi-portal | **CERRADO** | Login/logout, recuperación, sesión por portal. Registro obligatorio; usuario aislado por tenant; Google compatible con el aislamiento |
| RF13 | RBAC | **CERRADO** | 4 roles: SuperAdmin, Tenant Admin, Colaborador, Comprador. Cuenta única por tenant |
| RF14 | Gestión de tenants (CRUD) | **EN DEFINICIÓN** | Alta, baja lógica, edición, listado, cambio de estado. Ver preguntas abiertas en sección 11 |
| RF15 | Gestión de planes comerciales | PROPUESTO | Planes y qué módulos/límites incluye cada uno |
| RF16 | Configuración de comisiones | PROPUESTO | % por tenant y/o por evento |
| RF17 | Dashboard de facturación global | PROPUESTO | Métricas agregadas de ventas y comisiones |
| RF18 | Cupones y promociones | PROPUESTO | Descuento %/monto fijo, vigencia, límite de usos |
| RF19 | Gestión de colaboradores | PROPUESTO | Alta de colaboradores + asignación de eventos |
| RF20 | Reportería y exportación | PROPUESTO | Reportes de ventas/asistentes/ingresos, export CSV/Excel |
| RF21 | Gestión de activos visuales | PROPUESTO | Carga de logo, banner, favicon a S3 |
| RF22 | Ciclo de vida del evento | PROPUESTO | Estados borrador/publicado/finalizado |
| RF23 | Carrito y checkout | PROPUESTO | Selección múltiple, cupón, resumen de compra |
| RF24 | Captura de datos del comprador | PROPUESTO | Formulario que alimenta el CRM (RF11) |
| RF25 | Notificaciones transaccionales | PROPUESTO | Email de confirmación con PDF adjunto |

### 9.3. Requerimientos no funcionales — base del proyecto

| ID | Nombre | Descripción |
|---|---|---|
| RNF01 | Cero sobreventa | Bloqueo pesimista en BD para impedir compras simultáneas de la misma entrada |
| RNF02 | Tiempos de respuesta extremos | Resolución de dominio <5ms en Edge; validación QR <100ms |
| RNF03 | Aislamiento estricto de datos | Filtro obligatorio por tenant_id en cada consulta |
| RNF04 | Idempotencia en pagos | Procesamiento seguro de webhooks duplicados |
| RNF05 | Heterogeneidad estética | Dos tenants con mismos módulos pueden verse totalmente distintos, sin tocar código |

### 9.4. Requerimientos no funcionales — propuestos

| ID | Nombre | Descripción |
|---|---|---|
| RNF06 | Seguridad de aplicación | HTTPS, gestión de secretos, hashing, mitigación OWASP Top 10 |
| RNF07 | Cumplimiento en pagos | No almacenar datos de tarjeta; tokenización vía pasarela |
| RNF08 | Privacidad de datos personales | Cumplimiento normativo (Ley 25.326 AR / GDPR si aplica) |
| RNF09 | Disponibilidad | SLA objetivo (ej. 99.9%) con tolerancia a hot sales |
| RNF10 | Escalabilidad horizontal | Autoescalado de contenedores Go bajo demanda |
| RNF11 | Observabilidad | Logging estructurado, métricas, trazas, alertas |
| RNF12 | Respaldo y recuperación | Backups automáticos de PostgreSQL con RPO/RTO definidos |
| RNF13 | Auditoría | Registro inmutable de operaciones financieras y cambios críticos |
| RNF14 | Consistencia de caché | Estrategia de invalidación entre los 3 niveles de caché |

---

## 10. Alcance

### Dentro del alcance (Fase 1)
- Infraestructura BFF en Next.js con subdominios y dominios personalizados
- Backend en dos servicios Go (`ticketing-core`, `ticketing-platform`) con PostgreSQL y Redis
- Resolución de dominios en el Edge mediante Redis
- Portal SuperAdmin con gestor de tenants, dominios y editor visual (Puck)
- Portal Tenant Admin completo (eventos, entradas, precios, cupones, CRM)
- Sitio público adaptativo con motor de plantillas dinámico
- Galería de fotos y secciones modulares por plan
- Mercado Pago con comisión dividida + entradas QR
- API y cliente para validación de QR en puerta

### Fuera del alcance (fases futuras)
- Mercado secundario / reventa de entradas entre usuarios
- Selección de asientos en mapas 3D (Fase 1 solo cupos/sectores)
- Integración con impresoras térmicas de tickets físicos
- Mapeo de accesos con torniquetes o hardware dedicado

---

## 11. Decisiones pendientes

### Resueltas
- ✅ Pasarela de pagos: Mercado Pago (marketplace/Connect)
- ✅ Almacenamiento: AWS S3
- ✅ Arquitectura backend: dos servicios Go, base compartida, dueño de escritura por tabla
- ✅ Modelo de identidad: registro obligatorio, usuario por tenant, Google compatible
- ✅ Frontend: Next.js único para las 3 experiencias (no React SPA)
- ✅ Nombre del producto: GoAccess (`goaccess.com.ar`)
- ✅ Nombres de repos y dominios (sección 4.2 y 4.3)

### Aún por definir
- **Proveedor de autenticación:** JWT nativo en Go vs. proveedor externo (Clerk / Supabase Auth)
- **Mecanismo offline del escáner:** ¿necesita validar sin conexión? (afecta si el escáner es PWA simple o requiere más trabajo de almacenamiento local / app nativa)
- **Estructura interna de los repos Go:** Hexagonal vs. Clean Architecture; herramienta de migraciones (Goose, Atlas, golang-migrate)
- **Onboarding de Mercado Pago:** flujo para que cada tenant vincule su cuenta y reciba su parte del split
- **Formato final del escáner:** PWA dentro del mismo proyecto Next.js (opción recomendada, sin confirmar del todo) vs. app nativa separada — depende de la necesidad de modo offline

### En definición inmediata — RF14 (Gestión de tenants)
Preguntas abiertas:
1. Datos mínimos del tenant al crearlo (propuesta: nombre, slug/subdominio, email de contacto, plan, estado)
2. Qué ve el fan ante un tenant **suspendido**: ¿cartel de "tienda no disponible" o no carga directamente?
3. Al crear el tenant, ¿se crea también su primer Tenant Admin con email de invitación?

---

## 12. Roadmap de desarrollo (fases sugeridas)

1. **Fase 0 — Cimientos:** repos (`ticketing-core`, `ticketing-platform`, `ticketing-shared`), modelos GORM base, migraciones, Redis, middleware de resolución de dominio funcionando.
2. **Fase 1 — Identidad y RBAC:** auth, login por rol, aislamiento por tenant en cada query.
3. **Fase 2 — API de Plataforma mínima:** CRUD de tenants, planes, comisiones. Objetivo: poder crear un tenant de prueba.
4. **Fase 3 — Eventos e inventario:** CRUD de eventos + bloqueo pesimista anti-sobreventa (testeado con concurrencia real).
5. **Fase 4 — Checkout y pagos:** onboarding Mercado Pago, checkout, webhook idempotente, split payment.
6. **Fase 5 — Emisión y validación de tickets:** PDF + QR firmado, email de confirmación, endpoint de validación, app de escáner.
7. **Fase 6 — Tienda pública y diseño** (paralelizable desde temprano): Puck, tokens, variantes CVA, Theme Studio.
8. **Fase 7 — Preparación para producción:** seguridad, observabilidad, backups, auditoría, invalidación de caché, reportería.

**Recomendación:** no invertir tiempo temprano en el editor visual completo (Puck) — es la parte más grande de UI pero no bloquea nada técnico. Priorizar tener el flujo de compra end-to-end funcionando con estilos fijos, y recién después invertir en que sea configurable.

---

## 13. Convenciones de nombres

- **Repos backend:** prefijo `ticketing-` + sufijo de rol (`core`, `platform`, `shared`)
- **Dominios de API:** `api.{rol}.goaccess.com.ar`
- **Separador en nombres técnicos:** guion medio (`kebab-case`) para repos/dominios; guion bajo reservado para variables en código
- **Ejemplo de tenant de referencia usado en la documentación:** FiestaBresh (`fiestabresh.com`)
