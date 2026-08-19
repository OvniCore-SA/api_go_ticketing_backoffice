// Package producer_config implementa el patrón Decorator sobre la configuración
// efectiva de un Producer. La respuesta a la pregunta "¿qué ve realmente el
// tenant X?" se resuelve como una pila de ConfigLayer: cada capa toma la
// configuración parcial acumulada, la enriquece con su responsabilidad, y la
// devuelve para la siguiente.
//
// Este paquete NO reemplaza al CRUD de sub-recursos (producer_modules,
// design_tokens, page_templates, etc.). Solo agrega la vista de LECTURA
// consolidada que consumen ticketing-core y el panel del SuperAdmin.
//
// Ver docs/patrones.md (a crear) para el rationale del patrón.
package producer_config

import (
	"encoding/json"
	"time"
)

// EffectiveConfig es el estado final que "ve" un tenant en runtime,
// resultado de aplicar todas las capas de configuración en orden.
//
// Los campos jsonb son opacos para este servicio — se propagan tal cual;
// el frontend (Puck + CVA) los interpreta.
type EffectiveConfig struct {
	Meta     Meta                       `json:"meta"`
	Modules  []EffectiveModule          `json:"modules"`
	Variants map[uint]EffectiveVariant  `json:"variants"` // moduleID → variante elegida
	Tokens   EffectiveTokens            `json:"tokens"`
	Pages    map[string]json.RawMessage `json:"pages"` // pageType → puck_json
}

// Meta contiene información de trazabilidad — de qué capa vino cada dato y
// bandera globales (suspendido, plan, template). Facilita debugging y le da
// al frontend un lugar único de dónde decidir "modo pausa", etc.
type Meta struct {
	ProducerID   uint      `json:"producer_id"`
	ProducerSlug string    `json:"producer_slug"`
	ProducerName string    `json:"producer_name"`
	Status       string    `json:"status"`                  // active | suspended
	Suspended    bool      `json:"suspended"`               // shortcut para frontend
	TemplateID   *uint     `json:"template_id,omitempty"`
	TemplateCode string    `json:"template_code,omitempty"`
	PlanID       *uint     `json:"plan_id,omitempty"`
	PlanCode     string    `json:"plan_code,omitempty"`
	ResolvedAt   time.Time `json:"resolved_at"`
	AppliedLayers []string `json:"applied_layers"` // orden en que se aplicaron — útil para debug
}

// EffectiveModule representa un módulo del catálogo global tal como lo ve
// el tenant. Enabled=false significa "el sistema lo reconoce pero no debe
// renderizarse en la tienda pública".
type EffectiveModule struct {
	ID      uint   `json:"id"`
	Code    string `json:"code"`
	Name    string `json:"name"`
	IsCore  bool   `json:"is_core"`
	Enabled bool   `json:"enabled"`
	// Source indica qué capa dejó este módulo en el estado actual — útil
	// para debug cuando el SuperAdmin quiere entender "¿por qué está esto acá?".
	Source string `json:"source,omitempty"`
}

// EffectiveVariant es la variante visual elegida para un módulo.
type EffectiveVariant struct {
	ModuleID           uint   `json:"module_id"`
	ComponentVariantID uint   `json:"component_variant_id"`
	VariantCode        string `json:"variant_code,omitempty"`
}

// EffectiveTokens son los design tokens finales. Los cuatro campos son JSON
// opaco — colores, fuentes, radios y sombras — con la estructura que el
// frontend defina.
type EffectiveTokens struct {
	Colors  json.RawMessage `json:"colors,omitempty"`
	Fonts   json.RawMessage `json:"fonts,omitempty"`
	Radius  json.RawMessage `json:"radius,omitempty"`
	Shadows json.RawMessage `json:"shadows,omitempty"`
}

// newEmpty devuelve un EffectiveConfig inicializado y listo para que las
// capas lo enriquezcan. Se usa como estado inicial dentro del Resolver.
func newEmpty(producerID uint) EffectiveConfig {
	return EffectiveConfig{
		Meta:          Meta{ProducerID: producerID, AppliedLayers: []string{}},
		Modules:       []EffectiveModule{},
		Variants:      map[uint]EffectiveVariant{},
		Tokens:        EffectiveTokens{},
		Pages:         map[string]json.RawMessage{},
	}
}
