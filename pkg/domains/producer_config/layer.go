package producer_config

import (
	"context"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
)

// ResolveContext es el input compartido de todas las capas. Contiene los
// datos que el Resolver preloadeó una sola vez para evitar N+1 dentro de
// las capas — cada capa recibe el Producer con Template y Plan ya cargados.
// Si una capa necesita datos extra (ej. ProducerModules), los pide a su
// propio repositorio inyectado.
type ResolveContext struct {
	ProducerID uint
	Producer   entities.Producer
}

// ConfigLayer es una capa del Decorator. Recibe la configuración parcial
// acumulada por las capas previas, la enriquece con su responsabilidad,
// y la devuelve para que la siguiente la reciba.
//
// Contrato:
//   - No debe mutar el input: devolver una nueva EffectiveConfig.
//   - Debe ser idempotente: aplicarla dos veces produce el mismo resultado.
//   - Si no aplica en este producer (ej. plan sin restricciones), devuelve
//     la config sin tocarla — nunca error.
//   - Solo devuelve error para fallas de infraestructura (DB caída, etc.).
//     Datos faltantes no son error — la capa se salta silenciosamente.
type ConfigLayer interface {
	Name() string
	Apply(ctx context.Context, rctx ResolveContext, cfg EffectiveConfig) (EffectiveConfig, error)
}
