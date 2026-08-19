package producer_config

import (
	"context"
	"errors"
	"time"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producers"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// Resolver es el orquestador del patrón Decorator. Mantiene una lista de
// ConfigLayer en el orden en que deben aplicarse y expone una única API:
// Resolve(producerID) → EffectiveConfig.
//
// Agregar una nueva regla del negocio (ej. "durante campaña X activar
// banner promocional") se hace escribiendo una nueva ConfigLayer y
// enchufándola en el slice al construir el Resolver — no se toca código
// existente.
type Resolver struct {
	producersRepo producers.Repository
	layers        []ConfigLayer
	now           func() time.Time // inyectable para tests
}

func NewResolver(producersRepo producers.Repository, layers ...ConfigLayer) *Resolver {
	return &Resolver{
		producersRepo: producersRepo,
		layers:        layers,
		now:           time.Now,
	}
}

// Resolve compone la configuración efectiva del producer aplicando las
// capas en orden. Cada capa recibe la salida de la anterior; la primera
// recibe una EffectiveConfig vacía.
//
// Fallas:
//   - Producer inexistente → 404.
//   - Cualquier capa devuelve error → 500 con el nombre de la capa que
//     falló (facilita debugging en producción).
func (r *Resolver) Resolve(ctx context.Context, producerID uint) (EffectiveConfig, error) {
	if producerID == 0 {
		return EffectiveConfig{}, fiber.NewError(fiber.StatusBadRequest, "producer_id inválido")
	}

	producer, err := r.producersRepo.GetProducerRepository(filters.ProducerFilter{ID: producerID})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return EffectiveConfig{}, fiber.NewError(fiber.StatusNotFound, "producer no encontrado")
		}
		return EffectiveConfig{}, fiber.NewError(fiber.StatusInternalServerError, "error al obtener producer")
	}

	rctx := ResolveContext{
		ProducerID: producerID,
		Producer:   producer,
	}

	cfg := newEmpty(producerID)
	cfg.Meta.ProducerSlug = producer.Slug
	cfg.Meta.ProducerName = producer.Name
	cfg.Meta.Status = producer.Status
	cfg.Meta.TemplateID = producer.TemplateID
	cfg.Meta.PlanID = producer.PlanID
	cfg.Meta.ResolvedAt = r.now().UTC()

	for _, layer := range r.layers {
		next, err := layer.Apply(ctx, rctx, cfg)
		if err != nil {
			return EffectiveConfig{}, fiber.NewError(
				fiber.StatusInternalServerError,
				"error en capa "+layer.Name()+": "+err.Error(),
			)
		}
		cfg = next
		cfg.Meta.AppliedLayers = append(cfg.Meta.AppliedLayers, layer.Name())
	}

	return cfg, nil
}
