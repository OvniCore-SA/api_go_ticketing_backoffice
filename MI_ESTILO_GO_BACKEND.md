# Mi Estilo de Backend en Go — Guía Completa de Arquitectura y Patrones

> Este documento captura **exactamente** cómo programo APIs REST en Go. Es la referencia para replicar mi estilo en cualquier proyecto nuevo.

---

## Stack Tecnológico

| Herramienta | Uso |
|---|---|
| **Go 1.25.7** | Lenguaje (no actualizar) |
| **Fiber v2** | Framework HTTP |
| **GORM** | ORM |
| **PostgreSQL** | Base de datos |
| **golang-jwt/jwt v5** | Autenticación JWT |
| **godotenv** | Variables de entorno desde `.env` |
| **gomail v2** | Envío de emails SMTP |
| **google/uuid** | Generación de UUIDs |
| **bcrypt** | Hashing de contraseñas |

---

## Estructura de Carpetas

```
proyecto/
├── main.go                          # Entry point: solo llama api.SetupApp()
├── api/
│   ├── app.go                       # Bootstrap: DI manual, middlewares, rutas
│   ├── handlers/
│   │   ├── helpers.go               # handleServiceError() compartido
│   │   └── {domain}_handler.go
│   ├── middlewares/
│   │   └── auth.go                  # JWT middleware
│   └── routes/
│       └── {domain}.go
├── internal/
│   ├── config/
│   │   └── email.go                 # Configuración SMTP/gomail
│   ├── database/
│   │   └── postgres.go              # PostgresClient (wrapper de *gorm.DB)
│   ├── logs/
│   │   └── logs.go                  # Info, Error, Fatal
│   └── assets/                      # Archivos estáticos (logos, imágenes)
└── pkg/
    ├── commons/
    │   └── tools.go                 # Utilidades globales reutilizables
    ├── domains/
    │   └── {domain}/
    │       ├── {domain}_repository.go   # Interface Repository + implementación
    │       └── {domain}_service.go      # Interface Service + implementación
    ├── dtos/
    │   └── {domain}dtos/
    │       ├── request.go           # DTOs de entrada + Validate() + ToEntity()
    │       ├── response.go          # DTOs de salida + FromEntity() + FromEntities()
    │       └── errors.go            # Errores de dominio tipados
    ├── entities/
    │   └── {domain}.go              # Modelos GORM
    ├── filters/
    │   └── {domain}_filter.go       # Structs de filtro para queries
    └── dtos/utils/
        └── pagination.go            # Struct Pagination + NewPagination()
```

---

## 1. Entidades (GORM Models)

### Reglas
- Siempre embeben `gorm.Model` (agrega `ID`, `CreatedAt`, `UpdatedAt`, `DeletedAt` → soft delete automático).
- Campos nullable → puntero (`*uint`, `*float64`, `*time.Time`, `*string`).
- Relaciones opcionales → puntero en la FK y en el campo de asociación.
- Constantes de estado en el mismo archivo, arriba de la struct.
- Si el nombre de tabla no sigue la convención de GORM (pluralizado en snake_case del nombre de struct), se define `TableName()`.

### Patrón base

```go
package entities

import "gorm.io/gorm"

const (
    ClientStatusActive   = "active"
    ClientStatusInactive = "inactive"
    ClientStatusPending  = "pending"
)

type Client struct {
    gorm.Model
    BusinessName string         `gorm:"column:business_name;not null"`
    Sector       string         `gorm:"column:sector"`
    Status       string         `gorm:"column:status;default:'active'"`
    Batches      []VoucherBatch `gorm:"foreignKey:ClientID"`
}
```

### Con FK nullable y relación opcional

```go
type User struct {
    gorm.Model
    RoleID    uint     `gorm:"column:role_id;not null"`
    StationID *uint    `gorm:"column:station_id"`           // nullable → *uint
    FullName  string   `gorm:"column:full_name;not null"`
    Email     string   `gorm:"column:email;uniqueIndex;not null"`
    Status    string   `gorm:"column:status;default:'active'"`
    Role      Role     `gorm:"foreignKey:RoleID"`            // obligatoria
    Station   *Station `gorm:"foreignKey:StationID"`         // opcional → *Station
}
```

### Con FK explícita (campo custom "created_by")

```go
type VoucherBatch struct {
    gorm.Model
    ClientID      uint      `gorm:"column:client_id;not null"`
    CreatedBy     uint      `gorm:"column:created_by;not null"`     // FK manual
    UnitValue     *float64  `gorm:"column:unit_value"`               // nullable
    Status        string    `gorm:"column:status;default:'active'"`
    Client        Client    `gorm:"foreignKey:ClientID"`
    CreatedByUser User      `gorm:"foreignKey:CreatedBy"`            // nombre explicito
    Vouchers      []Voucher `gorm:"foreignKey:BatchID"`
}

// TableName cuando GORM no pluraliza correctamente
func (VoucherBatch) TableName() string {
    return "voucher_batches"
}
```

### Sin gorm.Model (entidades especiales)

```go
// Cuando no se necesita soft delete ni timestamps automáticos
type AuditLog struct {
    ID         int64           `gorm:"primaryKey;autoIncrement"`
    UserID     *uint           `gorm:"column:user_id"`
    Action     string          `gorm:"column:action;not null"`
    Details    json.RawMessage `gorm:"column:details;type:jsonb"`
    CreatedAt  time.Time
    User       *User `gorm:"foreignKey:UserID"`
}
```

### Many-to-many

```go
type Role struct {
    gorm.Model
    Code        string       `gorm:"column:code;uniqueIndex;not null"`
    Permissions []Permission `gorm:"many2many:role_permissions;"`
}
```

---

## 2. Filtros

Cada dominio tiene su propio filter struct en `pkg/filters/{domain}_filter.go`. Se usan para pasar condiciones de búsqueda al repositorio sin exponer parámetros sueltos.

```go
// pkg/filters/client_filter.go
package filters

type ClientFilter struct {
    ID uint
}
```

```go
// pkg/filters/voucher_filter.go
package filters

type VoucherFilter struct {
    ID           uint
    QRCode       string
    VoucherCode  string
    BatchID      uint
    Status       string
    SelectFields []string   // para columnas específicas con .Select()
}
```

```go
// pkg/filters/user_filter.go
package filters

type UserFilter struct {
    ID             uint
    Name           string
    Email          string
    SelectedFields []string
}
```

---

## 3. DTOs

### 3.1 Request — con `Validate()` y `ToEntity()`

El request DTO siempre tiene:
- Tags `json:"..."` para body, `query:"..."` para query params.
- Método `Validate() error` que valida con helpers de `commons`.
- Método `ToEntity()` cuando necesita convertirse a entidad GORM (no siempre en updates).

```go
// pkg/dtos/clientdtos/request.go
package clientdtos

import (
    "errors"
    "github.com/OvniCore-SA/api_go_vouchers/pkg/commons"
    "github.com/OvniCore-SA/api_go_vouchers/pkg/entities"
)

type RequestCreateClient struct {
    BusinessName string `json:"business_name"`
    Sector       string `json:"sector"`
}

func (r RequestCreateClient) Validate() error {
    if commons.StringIsEmpty(r.BusinessName) {
        return errors.New("la razón social es requerida")
    }
    return nil
}

func (r *RequestCreateClient) ToEntity() entities.Client {
    return entities.Client{
        BusinessName: r.BusinessName,
        Sector:       r.Sector,
        Status:       entities.ClientStatusActive,
    }
}

// Update: sin ToEntity() porque el servicio arma map[string]interface{}
type RequestUpdateClient struct {
    BusinessName string `json:"business_name"`
    Sector       string `json:"sector"`
    Status       string `json:"status"`
}

// List: query params
type RequestListClients struct {
    Search string `query:"search"`
    Status string `query:"status"`
    Page   int    `query:"page"`
    Limit  int    `query:"limit"`
}
```

### 3.2 Response — con `FromEntity()` y `FromEntities()`

- `FromEntity(e entities.X)` → mapea una entidad a un response DTO. Siempre con pointer receiver `*Response`.
- `FromEntities(list []entities.X)` → mapea un slice. Inicializa el slice interno con `make`.
- Relaciones opcionales (puntero en entidad) → puntero en el DTO, se mapean con `if e.Relation != nil`.
- Los campos `omitempty` en los tags se usan cuando el campo puede no estar presente.

```go
// pkg/dtos/clientdtos/response.go
package clientdtos

import (
    "time"
    "github.com/OvniCore-SA/api_go_vouchers/pkg/entities"
)

type ResponseClient struct {
    ID           uint      `json:"id"`
    BusinessName string    `json:"business_name"`
    Sector       string    `json:"sector"`
    Status       string    `json:"status"`
    CreatedAt    time.Time `json:"created_at"`
}

type ResponseClients struct {
    Clients []ResponseClient `json:"clients"`
}

func (r *ResponseClients) FromEntities(list []entities.Client) {
    r.Clients = make([]ResponseClient, len(list))
    for i, e := range list {
        r.Clients[i].FromEntity(e)
    }
}

func (r *ResponseClient) FromEntity(e entities.Client) {
    r.ID = e.ID
    r.BusinessName = e.BusinessName
    r.Sector = e.Sector
    r.Status = e.Status
    r.CreatedAt = e.CreatedAt
}
```

### Response con relaciones anidadas y punteros

```go
type ResponseVoucher struct {
    ID            uint                  `json:"id"`
    VoucherCode   string                `json:"voucher_code"`
    Status        string                `json:"status"`
    UsedAt        *time.Time            `json:"used_at,omitempty"`
    UsedBy        *ResponseUserShort    `json:"used_by,omitempty"`    // puntero opcional
    UsedAtStation *ResponseStationShort `json:"used_at_station,omitempty"`
}

func (r *ResponseVoucher) FromEntity(e entities.Voucher) {
    r.ID = e.ID
    r.VoucherCode = e.VoucherCode
    r.Status = e.Status
    r.UsedAt = e.UsedAt

    // Relaciones opcionales: mapear solo si la entidad las trajo precargadas
    if e.UsedByUser != nil {
        r.UsedBy = &ResponseUserShort{ID: e.UsedByUser.ID, FullName: e.UsedByUser.FullName}
    }
    if e.UsedAtStationObj != nil {
        r.UsedAtStation = &ResponseStationShort{
            ID:   e.UsedAtStationObj.ID,
            Code: e.UsedAtStationObj.Code,
            Name: e.UsedAtStationObj.Name,
        }
    }
}
```

### Response con relación obligatoria (sub-DTO inline)

```go
// Sub-DTOs "Short" para relaciones embebidas (no exponen todo el objeto)
type ResponseClientShort struct {
    ID           uint   `json:"id"`
    BusinessName string `json:"business_name"`
}

type ResponseBatch struct {
    ID      uint                `json:"id"`
    Client  ResponseClientShort `json:"client"`    // embebido, no puntero
    Product ResponseProductShort `json:"product"`
}

func (r *ResponseBatch) FromEntity(e entities.VoucherBatch) {
    r.ID = e.ID
    r.Client = ResponseClientShort{
        ID:           e.Client.ID,
        BusinessName: e.Client.BusinessName,
    }
    r.Product = ResponseProductShort{
        ID:   e.Product.ID,
        Name: e.Product.Name,
    }
}
```

### 3.3 Errors — variables exportadas

```go
// pkg/dtos/clientdtos/errors.go
package clientdtos

import "errors"

var (
    ErrClientNotFound      = errors.New("cliente no encontrado")
    ErrClientHasActiveBatches = errors.New("el cliente tiene lotes activos")
)
```

---

## 4. Repositorio

### Estructura

Siempre en el mismo archivo `{domain}_repository.go`. Primero la interface, luego el struct privado, luego el constructor, luego los métodos.

```go
package clients

import (
    "github.com/OvniCore-SA/api_go_vouchers/internal/database"
    "github.com/OvniCore-SA/api_go_vouchers/pkg/entities"
    "github.com/OvniCore-SA/api_go_vouchers/pkg/filters"
)

type Repository interface {
    CreateClientRepository(entity entities.Client) (entities.Client, error)
    GetAllClientsRepository(search, status string, page, limit int) ([]entities.Client, int64, error)
    GetClientRepository(filter filters.ClientFilter) (entities.Client, error)
    UpdateClientRepository(id uint, fields map[string]interface{}) error
    DeleteClientRepository(id uint) error
}

type repository struct {
    db *database.PostgresClient
}

func NewClientsRepository(db *database.PostgresClient) Repository {
    return &repository{db: db}
}
```

### Create

```go
func (r *repository) CreateClientRepository(entity entities.Client) (entities.Client, error) {
    if err := r.db.Create(&entity).Error; err != nil {
        return entities.Client{}, err
    }
    return entity, nil
}
```

### GetAll con filtros opcionales + paginación

```go
func (r *repository) GetAllClientsRepository(search, status string, page, limit int) ([]entities.Client, int64, error) {
    var list []entities.Client
    var total int64

    query := r.db.Model(&entities.Client{})

    if search != "" {
        query = query.Where("business_name ILIKE ?", "%"+search+"%")
    }
    if status != "" {
        query = query.Where("status = ?", status)
    }

    query.Count(&total)

    if err := query.Order("business_name ASC").
        Offset((page - 1) * limit).
        Limit(limit).
        Find(&list).Error; err != nil {
        return nil, 0, err
    }
    return list, total, nil
}
```

### GetOne con filtro struct + preloads

```go
func (r *repository) GetBatchByIDRepository(id uint) (entities.VoucherBatch, error) {
    var batch entities.VoucherBatch
    if err := r.db.
        Preload("Client").
        Preload("Product").
        Preload("CreatedByUser").
        Where("id = ?", id).
        First(&batch).Error; err != nil {
        return entities.VoucherBatch{}, err
    }
    return batch, nil
}
```

### GetOne con filter struct dinámico

```go
func (r *repository) GetVoucherRepository(filter filters.VoucherFilter) (entities.Voucher, error) {
    var entity entities.Voucher
    query := r.db.Model(&entities.Voucher{}).
        Preload("Batch.Client").
        Preload("Batch.Product").
        Preload("UsedByUser").
        Preload("UsedAtStationObj")

    if filter.ID > 0 {
        query = query.Where("id = ?", filter.ID)
    }
    if filter.QRCode != "" {
        query = query.Where("qr_code = ?", filter.QRCode)
    }
    if filter.VoucherCode != "" {
        query = query.Where("voucher_code = ?", filter.VoucherCode)
    }
    if len(filter.SelectFields) > 0 {
        query = query.Select(filter.SelectFields)
    }

    err := query.First(&entity).Error
    return entity, err
}
```

### Update con `map[string]interface{}`

Siempre se actualiza con un mapa de campos, nunca guardando la entidad completa. Esto evita que GORM actualice campos con zero-value no deseados.

```go
func (r *repository) UpdateClientRepository(id uint, fields map[string]interface{}) error {
    return r.db.Model(&entities.Client{}).Where("id = ?", id).Updates(fields).Error
}
```

### Soft Delete

```go
func (r *repository) DeleteClientRepository(id uint) error {
    return r.db.Delete(&entities.Client{}, id).Error  // soft delete por gorm.Model
}
```

### Transacción (Create masivo en dos tablas)

```go
func (r *repository) CreateBatchWithVouchersRepository(batch entities.VoucherBatch, voucherList []entities.Voucher) (entities.VoucherBatch, error) {
    tx := r.db.Begin()
    if tx.Error != nil {
        return entities.VoucherBatch{}, tx.Error
    }
    defer tx.Rollback()

    if err := tx.Create(&batch).Error; err != nil {
        return entities.VoucherBatch{}, err
    }

    for i := range voucherList {
        voucherList[i].BatchID = batch.ID
    }

    if err := tx.CreateInBatches(&voucherList, 500).Error; err != nil {
        return entities.VoucherBatch{}, err
    }

    if err := tx.Commit().Error; err != nil {
        return entities.VoucherBatch{}, err
    }
    return batch, nil
}
```

### Aggregate / Count por grupo

```go
func (r *repository) CountBatchVouchersByStatusRepository(batchID uint) (map[string]int, error) {
    type result struct {
        Status string
        Count  int
    }
    var rows []result
    err := r.db.Model(&entities.Voucher{}).
        Select("status, COUNT(*) as count").
        Where("batch_id = ?", batchID).
        Group("status").
        Scan(&rows).Error
    if err != nil {
        return nil, err
    }
    counts := map[string]int{}
    for _, row := range rows {
        counts[row.Status] = row.Count
    }
    return counts, nil
}
```

### Pluck (extraer un solo campo)

```go
func (r *repository) GetLastBatchCodeRepository() (string, error) {
    var code string
    err := r.db.Model(&entities.VoucherBatch{}).
        Select("batch_code").
        Order("id DESC").
        Limit(1).
        Pluck("batch_code", &code).Error
    return code, err
}
```

### Update masivo con condición

```go
func (r *repository) UpdateVouchersInBatchRepository(batchID uint, whereStatus string, fields map[string]interface{}) (int64, error) {
    result := r.db.Model(&entities.Voucher{}).
        Where("batch_id = ? AND status = ?", batchID, whereStatus).
        Updates(fields)
    return result.RowsAffected, result.Error
}
```

---

## 5. Servicio

### Estructura

Siempre en `{domain}_service.go`. Interface primero, struct privado, constructor, métodos.

```go
package clients

import (
    "github.com/OvniCore-SA/api_go_vouchers/pkg/dtos/clientdtos"
    "github.com/OvniCore-SA/api_go_vouchers/pkg/dtos/utils"
    "github.com/OvniCore-SA/api_go_vouchers/pkg/filters"
    "github.com/gofiber/fiber/v2"
)

type Service interface {
    CreateClientService(req clientdtos.RequestCreateClient) (clientdtos.ResponseClient, error)
    GetClientsService(req clientdtos.RequestListClients) (clientdtos.ResponseClients, utils.Pagination, error)
    GetClientByIDService(id uint) (clientdtos.ResponseClient, error)
    UpdateClientService(id uint, req clientdtos.RequestUpdateClient) (clientdtos.ResponseClient, error)
    DeleteClientService(id uint) error
}

type service struct {
    repo Repository
}

func NewClientsService(repo Repository) Service {
    return &service{repo: repo}
}
```

### Create

```go
func (s *service) CreateClientService(req clientdtos.RequestCreateClient) (clientdtos.ResponseClient, error) {
    if err := req.Validate(); err != nil {
        return clientdtos.ResponseClient{}, fiber.NewError(fiber.StatusBadRequest, err.Error())
    }

    created, err := s.repo.CreateClientRepository(req.ToEntity())
    if err != nil {
        return clientdtos.ResponseClient{}, fiber.NewError(fiber.StatusInternalServerError, "error al crear el cliente")
    }

    var response clientdtos.ResponseClient
    response.FromEntity(created)
    return response, nil
}
```

### GetAll con paginación

```go
func (s *service) GetClientsService(req clientdtos.RequestListClients) (response clientdtos.ResponseClients, pagination utils.Pagination, err error) {
    if req.Page < 1 {
        req.Page = 1
    }
    if req.Limit < 1 {
        req.Limit = 20
    }

    list, total, err := s.repo.GetAllClientsRepository(req.Search, req.Status, req.Page, req.Limit)
    if err != nil {
        err = fiber.NewError(fiber.StatusInternalServerError, "error al obtener clientes")
        return
    }

    response.FromEntities(list)
    pagination = utils.NewPagination(req.Page, req.Limit, total)
    return
}
```

### GetByID

```go
func (s *service) GetClientByIDService(id uint) (clientdtos.ResponseClient, error) {
    client, err := s.repo.GetClientRepository(filters.ClientFilter{ID: id})
    if err != nil {
        return clientdtos.ResponseClient{}, fiber.NewError(fiber.StatusNotFound, clientdtos.ErrClientNotFound.Error())
    }
    var response clientdtos.ResponseClient
    response.FromEntity(client)
    return response, nil
}
```

### Update con mapa de campos dinámico

```go
func (s *service) UpdateClientService(id uint, req clientdtos.RequestUpdateClient) (clientdtos.ResponseClient, error) {
    // 1. Verificar existencia
    if _, err := s.repo.GetClientRepository(filters.ClientFilter{ID: id}); err != nil {
        return clientdtos.ResponseClient{}, fiber.NewError(fiber.StatusNotFound, clientdtos.ErrClientNotFound.Error())
    }

    // 2. Construir mapa con solo los campos no vacíos
    fields := map[string]interface{}{}
    if req.BusinessName != "" {
        fields["business_name"] = req.BusinessName
    }
    if req.Sector != "" {
        fields["sector"] = req.Sector
    }
    if req.Status != "" {
        fields["status"] = req.Status
    }

    // 3. Aplicar update
    if err := s.repo.UpdateClientRepository(id, fields); err != nil {
        return clientdtos.ResponseClient{}, fiber.NewError(fiber.StatusInternalServerError, "error al actualizar el cliente")
    }

    // 4. Re-fetch y retornar el estado actualizado
    updated, _ := s.repo.GetClientRepository(filters.ClientFilter{ID: id})
    var response clientdtos.ResponseClient
    response.FromEntity(updated)
    return response, nil
}
```

### Delete con validación de integridad

```go
func (s *service) DeleteClientService(id uint) error {
    if _, err := s.repo.GetClientRepository(filters.ClientFilter{ID: id}); err != nil {
        return fiber.NewError(fiber.StatusNotFound, clientdtos.ErrClientNotFound.Error())
    }

    count, err := s.repo.CountActiveBatchesByClientRepository(id)
    if err == nil && count > 0 {
        return fiber.NewError(fiber.StatusConflict, clientdtos.ErrClientHasActiveBatches.Error())
    }

    return s.repo.DeleteClientRepository(id)
}
```

### Errores en servicios

Siempre `fiber.NewError(fiber.StatusXxx, "mensaje")`. El mensaje es en español y legible para el usuario final.

```go
// 400 Bad Request → validación de input
fiber.NewError(fiber.StatusBadRequest, err.Error())

// 404 Not Found → recurso no existe
fiber.NewError(fiber.StatusNotFound, clientdtos.ErrClientNotFound.Error())

// 409 Conflict → regla de negocio viola estado actual
fiber.NewError(fiber.StatusConflict, "el lote ya fue anulado")

// 422 Unprocessable Entity → recurso existe pero no puede usarse
fiber.NewError(fiber.StatusUnprocessableEntity, "el cliente no está activo")

// 500 Internal Server Error → fallo de DB u otro sistema
fiber.NewError(fiber.StatusInternalServerError, "error al crear el cliente")
```

### Servicio con dependencias cruzadas

Cuando un servicio necesita validar datos de otro dominio, recibe los repositorios de esos dominios en el constructor:

```go
type service struct {
    repo         Repository
    clientsRepo  clients.Repository    // repositorio de otro dominio
    productsRepo products.Repository
    pdfSvc       PDFService
}

func NewBatchesService(repo Repository, clientsRepo clients.Repository, productsRepo products.Repository, pdfSvc PDFService) Service {
    return &service{
        repo:         repo,
        clientsRepo:  clientsRepo,
        productsRepo: productsRepo,
        pdfSvc:       pdfSvc,
    }
}
```

### Operaciones asíncronas (goroutines)

Para operaciones que no deben bloquear la respuesta (emails, audit logs):

```go
go func() {
    err := s.util.SendMailService(emailRequest)
    if err != nil {
        fmt.Printf("error al enviar el email: %s", err.Error())
    }
}()
```

---

## 6. Handlers

### Estructura

```go
package handlers

import (
    "strconv"
    "github.com/OvniCore-SA/api_go_vouchers/pkg/domains/clients"
    "github.com/OvniCore-SA/api_go_vouchers/pkg/dtos/clientdtos"
    "github.com/gofiber/fiber/v2"
)

type ClientsHandler struct {
    service clients.Service
}

func NewClientsHandler(service clients.Service) *ClientsHandler {
    return &ClientsHandler{service: service}
}
```

### Create (POST body)

```go
func (h *ClientsHandler) CreateClient(c *fiber.Ctx) error {
    var req clientdtos.RequestCreateClient
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
    }
    result, err := h.service.CreateClientService(req)
    if err != nil {
        return handleServiceError(c, err)
    }
    return c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": true, "data": result})
}
```

### GetAll (GET con query params + meta de paginación)

```go
func (h *ClientsHandler) GetClients(c *fiber.Ctx) error {
    var req clientdtos.RequestListClients
    if err := c.QueryParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "query params inválidos"})
    }
    list, meta, err := h.service.GetClientsService(req)
    if err != nil {
        return handleServiceError(c, err)
    }
    return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": list, "meta": meta})
}
```

### GetByID (parseo de :id)

```go
func (h *ClientsHandler) GetClientByID(c *fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "ID inválido"})
    }
    result, err := h.service.GetClientByIDService(uint(id))
    if err != nil {
        return handleServiceError(c, err)
    }
    return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}
```

### Update (PUT)

```go
func (h *ClientsHandler) UpdateClient(c *fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "ID inválido"})
    }
    var req clientdtos.RequestUpdateClient
    if err := c.BodyParser(&req); err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})
    }
    result, err := h.service.UpdateClientService(uint(id), req)
    if err != nil {
        return handleServiceError(c, err)
    }
    return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})
}
```

### Delete

```go
func (h *ClientsHandler) DeleteClient(c *fiber.Ctx) error {
    id, err := strconv.ParseUint(c.Params("id"), 10, 64)
    if err != nil {
        return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "ID inválido"})
    }
    if err := h.service.DeleteClientService(uint(id)); err != nil {
        return handleServiceError(c, err)
    }
    return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "message": "Cliente eliminado"})
}
```

### Leer el usuario autenticado del contexto

El middleware de auth inyecta el usuario en `c.Locals("user")`. Los handlers lo recuperan así:

```go
func (h *BatchesHandler) CreateBatch(c *fiber.Ctx) error {
    caller := c.Locals("user").(authdtos.ResponseUser)  // type assertion
    // usar caller.ID, caller.RoleCode, caller.StationID
}
```

### Respuesta con archivo binario (PDF)

```go
func (h *BatchesHandler) ExportPDF(c *fiber.Ctx) error {
    id, _ := strconv.ParseUint(c.Params("id"), 10, 64)
    pdfBytes, err := h.service.ExportBatchPDFService(uint(id), false)
    if err != nil {
        return handleServiceError(c, err)
    }
    c.Set("Content-Type", "application/pdf")
    c.Set("Content-Disposition", `attachment; filename="vouchers_lote_`+c.Params("id")+`.pdf"`)
    return c.Send(pdfBytes)
}
```

### helpers.go — manejo centralizado de errores de servicio

```go
// api/handlers/helpers.go
package handlers

import "github.com/gofiber/fiber/v2"

func handleServiceError(c *fiber.Ctx, err error) error {
    if fiberErr, ok := err.(*fiber.Error); ok {
        return c.Status(fiberErr.Code).JSON(fiber.Map{
            "status":  false,
            "message": fiberErr.Message,
        })
    }
    return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
        "status":  false,
        "message": err.Error(),
    })
}
```

---

## 7. Routes

```go
// api/routes/clients.go
package routes

import (
    "github.com/OvniCore-SA/api_go_vouchers/api/handlers"
    "github.com/gofiber/fiber/v2"
)

func SetupClientRoutes(router fiber.Router, handler *handlers.ClientsHandler, auth fiber.Handler) {
    clients := router.Group("/clients", auth)   // auth aplicado al grupo completo
    clients.Get("/", handler.GetClients)
    clients.Get("/:id", handler.GetClientByID)
    clients.Post("/", handler.CreateClient)
    clients.Put("/:id", handler.UpdateClient)
    clients.Delete("/:id", handler.DeleteClient)
}
```

Cuando hay rutas públicas y protegidas mezcladas:

```go
func SetupAuthRoutes(router fiber.Router, handler *handlers.AuthHandler, authMiddleware fiber.Handler) {
    auth := router.Group("/auth")

    // Rutas públicas (sin middleware)
    auth.Post("/login", handler.Login)
    auth.Post("/register", handler.Register)
    auth.Post("/restore-password", handler.RestorePassword)

    // Rutas protegidas (con middleware)
    protected := auth.Group("", authMiddleware)
    protected.Get("/me", handler.Me)
    protected.Post("/refresh", handler.Refresh)
    protected.Put("/change-password", handler.ChangePassword)
}
```

---

## 8. Bootstrap — `api/app.go`

Todo el wiring manual en `SetupApp()`. Orden fijo:

```go
func SetupApp() *fiber.App {
    // 1. Infraestructura
    db := database.NewPostgresClient()

    // 2. Repositorios
    clientsRepo := clients.NewClientsRepository(db)
    batchesRepo := batches.NewBatchesRepository(db)

    // 3. Servicios (componen repositorios, y repos externos si aplica)
    clientsService := clients.NewClientsService(clientsRepo)
    batchesService := batches.NewBatchesService(batchesRepo, clientsRepo, productsRepo, pdfSvc)

    // 4. Middleware de autenticación
    mw := middlewares.NewMiddlewareManager(&authService)
    authMiddleware := mw.ValidateToken()

    // 5. Handlers
    clientsHandler := handlers.NewClientsHandler(clientsService)

    // 6. Fiber app + middlewares globales
    app := fiber.New(fiber.Config{BodyLimit: 20 * 1024 * 1024})
    app.Use(logger.New())
    app.Use(recover.New())
    app.Use(cors.New(cors.Config{
        AllowOrigins: os.Getenv("CORS_ALLOW_ORIGINS"),
        AllowHeaders: "Content-Type, Authorization, Accept, Cache-Control",
        AllowMethods: "GET,POST,PUT,DELETE",
    }))

    // 7. Rutas versionadas
    apiVersion := os.Getenv("API_VERSION")
    if apiVersion == "" {
        apiVersion = "v1"
    }
    apiv1 := app.Group("/api/" + apiVersion)
    routes.SetupClientRoutes(apiv1, clientsHandler, authMiddleware)

    // 8. Servidor
    port := os.Getenv("APP_PORT")
    if port == "" {
        port = "8888"
    }
    app.Listen(":" + port)
    return app
}
```

---

## 9. Auth Middleware

JWT HS256. El middleware valida el token, busca el usuario en DB y lo inyecta en `c.Locals("user")` como `authdtos.ResponseUser`.

```go
// api/middlewares/auth.go
func (m *MiddlewareManager) ValidateToken() func(c *fiber.Ctx) error {
    return func(c *fiber.Ctx) error {
        bearer := c.Get("Authorization")
        parts := strings.Split(bearer, " ")
        if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
            return fiber.NewError(fiber.StatusUnauthorized, "formato de token inválido")
        }

        token, err := jwt.Parse(parts[1], func(token *jwt.Token) (interface{}, error) {
            return []byte(os.Getenv("JWT_SECRET_KEY")), nil
        })
        if err != nil || !token.Valid {
            return fiber.NewError(fiber.StatusUnauthorized, "token inválido")
        }

        claims := token.Claims.(jwt.MapClaims)
        userID, _ := strconv.ParseUint(claims["sub"].(string), 10, 64)

        user, err := (*m.authService).GetUserService(filters.UserFilter{ID: uint(userID)})
        if err != nil {
            return fiber.NewError(fiber.StatusInternalServerError, "error al obtener el usuario")
        }

        c.Locals("user", user)   // disponible en handlers como c.Locals("user").(authdtos.ResponseUser)
        return c.Next()
    }
}
```

### Generación de tokens (48h access, 5d refresh)

```go
claims := jwt.MapClaims{
    "iss":  "api_go_vouchers",
    "sub":  fmt.Sprintf("%d", user.ID),
    "user": map[string]interface{}{
        "id":         user.ID,
        "email":      user.Email,
        "role":       user.Role.Code,
        "station_id": user.StationID,
    },
    "exp": time.Now().Add(48 * time.Hour).Unix(),
}
token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
signedToken, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
```

---

## 10. Paginación

### En el repositorio

```go
// Contar total antes de paginar
query.Count(&total)

// Paginar
query.Offset((page - 1) * limit).Limit(limit).Find(&list)
```

### En el servicio

```go
pagination = utils.NewPagination(req.Page, req.Limit, total)
```

### Struct y constructor

```go
// pkg/dtos/utils/pagination.go
type Pagination struct {
    Page       int   `json:"page"`
    Limit      int   `json:"limit"`
    Total      int64 `json:"total"`
    TotalPages int   `json:"total_pages"`
}

func NewPagination(page, limit int, total int64) Pagination {
    if page < 1 { page = 1 }
    if limit < 1 { limit = 20 }
    if limit > 100 { limit = 100 }
    totalPages := int(total) / limit
    if int(total)%limit > 0 { totalPages++ }
    return Pagination{Page: page, Limit: limit, Total: total, TotalPages: totalPages}
}
```

### En el handler

```go
return c.JSON(fiber.Map{"status": true, "data": list, "meta": meta})
```

---

## 11. Formato de Respuesta HTTP

```go
// Éxito con data
c.Status(fiber.StatusCreated).JSON(fiber.Map{"status": true, "data": result})
c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": result})

// Éxito con data + paginación
c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "data": list, "meta": meta})

// Éxito sin data (delete)
c.Status(fiber.StatusOK).JSON(fiber.Map{"status": true, "message": "Cliente eliminado"})

// Error de parseo (siempre en el handler, no va al servicio)
c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": "payload inválido"})

// Error de servicio (delegar a handleServiceError)
return handleServiceError(c, err)
```

---

## 12. Commons — `pkg/commons/tools.go`

Utilidades globales disponibles en cualquier capa.

| Función | Uso |
|---|---|
| `StringIsEmpty(s string) bool` | Verifica si un string es vacío/solo espacios |
| `IsEmailValid(e string) bool` | Valida formato de email con regex |
| `IsNameValid(name string) error` | Al menos 3 letras, solo letras y espacios |
| `ValidatePassword(p string) error` | Min 8 chars, al menos un número, sin espacios |
| `HashPassword(p string) (string, error)` | bcrypt con costo 14 |
| `CheckPasswordHash(p, hash string) bool` | Compara password con hash |
| `GenerateToken() (string, error)` | Token seguro hex de 32 bytes (para email/reset) |
| `GenerateUUID() string` | UUID v4 como string |
| `GenerateVoucherCode() (string, error)` | 6 chars aleatorios del charset `ABCDEFGHJKLMNPQRSTUVWXYZ23456789` |
| `IsUUIDValid(u string) bool` | Verifica si un string es UUID válido |
| `TimeNowArgentina() time.Time` | `time.Now().UTC()` |
| `GetDateFirstMomentTime(t) time.Time` | 00:00:00 del día |
| `GetDateLastMomentTime(t) time.Time` | 23:59:59 del día |
| `FormatNombre(name string) string` | Primera letra en mayúscula |
| `StringInSlice(s string, slice []string) bool` | Pertenencia en slice |

---

## 13. PostgresClient

```go
// internal/database/postgres.go
type PostgresClient struct {
    *gorm.DB
}

func NewPostgresClient() *PostgresClient {
    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable search_path=%s TimeZone=UTC",
        os.Getenv("DB_HOST"), os.Getenv("DB_PORT"), os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_SCHEMA"),
    )
    gormDB, _ := gorm.Open(postgres.Open(dsn), &gorm.Config{
        NowFunc: func() time.Time { return time.Now().UTC() },
    })
    db, _ := gormDB.DB()
    db.SetMaxIdleConns(20)
    db.SetMaxOpenConns(50)
    db.SetConnMaxLifetime(time.Hour)
    return &PostgresClient{gormDB}
}
```

Los repositorios reciben `*database.PostgresClient`. El auth repository es la única excepción: recibe `*gorm.DB` directamente.

---

## 14. Convenciones de Nomenclatura

| Elemento | Convención | Ejemplo |
|---|---|---|
| Archivos | `snake_case` | `voucher_batch.go`, `clients_service.go` |
| Paquetes | lowercase singular | `package clients`, `package batches` |
| Interface de servicio | `Service` | `type Service interface` |
| Interface de repositorio | `Repository` | `type Repository interface` |
| Interface de auth | Prefijada con I | `type IAuthService interface` |
| Struct privado (impl) | camelCase | `type service struct`, `type repository struct` |
| Constructores | `New{Type}` | `NewClientsService`, `NewBatchesRepository` |
| Métodos de servicio | `{Acción}{Dominio}Service` | `CreateClientService`, `GetBatchByIDService` |
| Métodos de repositorio | `{Acción}{Dominio}Repository` | `CreateClientRepository`, `UpdateVoucherRepository` |
| Request DTOs | `Request{Acción}` | `RequestCreateClient`, `RequestListBatches` |
| Response DTOs | `Response{Dominio}` | `ResponseClient`, `ResponseBatch` |
| Response plural | `Response{Dominio}s` | `ResponseClients`, `ResponseBatches` |
| Errores de dominio | `Err{Descripción}` | `ErrClientNotFound`, `ErrBatchAlreadyAnnulled` |
| Constantes de estado | `{Dominio}Status{Estado}` | `ClientStatusActive`, `VoucherStatusUsed` |
| Constantes de código | `{Dominio}{Tipo}{Valor}` | `ProductCategoryFuel`, `UnitTypeLiters` |
| Handlers | `{Dominio}Handler` | `ClientsHandler`, `BatchesHandler` |
| Setup de rutas | `Setup{Dominio}Routes` | `SetupClientRoutes`, `SetupAuthRoutes` |

---

## 15. Checklist para Nuevo Dominio

```
[ ] pkg/entities/{domain}.go
      - gorm.Model embebido
      - Constantes de estado arriba de la struct
      - TableName() si el nombre no es obvio
      - FKs nullables como *uint
      - Relaciones opcionales como puntero

[ ] pkg/filters/{domain}_filter.go
      - Campos para cada condición de búsqueda posible
      - SelectFields []string si se necesita select parcial

[ ] pkg/dtos/{domain}dtos/errors.go
      - var ErrXxx = errors.New("mensaje en español")

[ ] pkg/dtos/{domain}dtos/request.go
      - RequestCreate con Validate() y ToEntity()
      - RequestUpdate (sin ToEntity, solo campos opcionales)
      - RequestList con tags query:"..." y Page/Limit int

[ ] pkg/dtos/{domain}dtos/response.go
      - ResponseX con FromEntity(*pointer receiver)
      - ResponseXs con FromEntities(*pointer receiver, make interno)
      - Sub-DTOs "Short" para relaciones embebidas
      - Punteros para relaciones opcionales + omitempty

[ ] pkg/domains/{domain}/{domain}_repository.go
      - Interface Repository + struct repository + constructor
      - CreateXRepository → recibe entidad, retorna entidad
      - GetAllXRepository → recibe filtros sueltos + page/limit, retorna slice + int64
      - GetXRepository → recibe filter struct, retorna entidad
      - UpdateXRepository → recibe id + map[string]interface{}, retorna error
      - DeleteXRepository → recibe id, retorna error

[ ] pkg/domains/{domain}/{domain}_service.go
      - Interface Service + struct service + constructor
      - Validate() antes de cualquier operación en Create
      - fiber.NewError() para todos los errores
      - map[string]interface{} para updates parciales
      - Re-fetch después de Update para retornar estado actualizado

[ ] api/handlers/{domain}_handler.go
      - struct {Domain}Handler con service
      - c.BodyParser para POST/PUT
      - c.QueryParser para GET con filtros
      - strconv.ParseUint para :id params
      - handleServiceError(c, err) para delegar errores
      - c.Locals("user").(authdtos.ResponseUser) cuando necesita el caller

[ ] api/routes/{domain}.go
      - func Setup{Domain}Routes(router fiber.Router, handler *handlers.{Domain}Handler, auth fiber.Handler)
      - router.Group("/{domain}", auth) para rutas todas protegidas
      - Separar en protected/public si hay mezcla

[ ] api/app.go
      - Agregar repo (paso 2)
      - Agregar service (paso 3)
      - Agregar handler (paso 5)
      - Llamar Setup{Domain}Routes (paso 7)
```
