package producer_config

import (
	"context"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
)

// StatusLayer aplica las reglas asociadas al Producer.Status. Se corre
// última para tener la última palabra: si el tenant está suspendido, esta
// capa marca la config para que el frontend renderice el modo pausa
// independientemente de lo que hayan dejado las capas anteriores.
//
// Reglas actuales:
//   - active: no toca nada.
//   - suspended: setea Meta.Suspended=true. La decisión final de "qué
//     renderizar" queda al frontend (ej. mostrar cartel "tienda pausada"
//     o degradar a solo página de contacto). NO se borran datos del
//     resto de la config — así el frontend puede seguir mostrando marca,
//     tokens, etc., pero saber que las compras están cerradas.
type StatusLayer struct{}

func NewStatusLayer() *StatusLayer { return &StatusLayer{} }

func (l *StatusLayer) Name() string { return "status" }

func (l *StatusLayer) Apply(ctx context.Context, rctx ResolveContext, cfg EffectiveConfig) (EffectiveConfig, error) {
	if rctx.Producer.Status == entities.ProducerStatusSuspended {
		cfg.Meta.Suspended = true
	}
	return cfg, nil
}
