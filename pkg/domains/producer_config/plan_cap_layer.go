package producer_config

import "context"

// PlanCapLayer restringe los módulos habilitados según el plan del producer.
//
// Estado actual: NO-OP. Está montada en el pipeline para que, cuando se
// implemente la tabla `plan_modules` (feature flags reales por plan), solo
// haya que sustituir la implementación de Apply — el orden del Resolver y
// las capas siguientes no cambian.
//
// Cuando se active:
//   1. Si el producer no tiene PlanID → no-op.
//   2. Cargar plan_modules del plan del producer.
//   3. Para cada EffectiveModule cuyo code no esté en plan_modules:
//      - Si IsCore=true → mantenerlo enabled (los core no se pueden capar).
//      - Si IsCore=false → forzar enabled=false + Source="plan_cap".
type PlanCapLayer struct{}

func NewPlanCapLayer() *PlanCapLayer { return &PlanCapLayer{} }

func (l *PlanCapLayer) Name() string { return "plan_cap" }

func (l *PlanCapLayer) Apply(ctx context.Context, rctx ResolveContext, cfg EffectiveConfig) (EffectiveConfig, error) {
	if rctx.Producer.Plan != nil {
		cfg.Meta.PlanCode = rctx.Producer.Plan.Code
	}
	// No-op por ahora — ver comentario del struct.
	return cfg, nil
}
