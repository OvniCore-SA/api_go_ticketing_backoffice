package producer_config

import (
	"context"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/templates"
)

// TemplateDefaultsLayer aplica los valores por defecto de la Template
// asociada al Producer. Sirve como fallback DEFENSIVO: si por algún motivo
// (bug de seed, migración, alta parcial) el producer no tiene sub-recursos,
// al menos hereda los valores de su template.
//
// En un producer sano, la capa siguiente (ProducerOverridesLayer) va a
// reemplazar todo lo que puso esta capa — es fuente única de verdad.
// TemplateDefaultsLayer existe para que "leer" nunca devuelva una config
// vacía por culpa de un dato faltante.
//
// Si el producer NO tiene template, esta capa es no-op.
type TemplateDefaultsLayer struct {
	templatesRepo templates.Repository
}

func NewTemplateDefaultsLayer(templatesRepo templates.Repository) *TemplateDefaultsLayer {
	return &TemplateDefaultsLayer{templatesRepo: templatesRepo}
}

func (l *TemplateDefaultsLayer) Name() string { return "template_defaults" }

func (l *TemplateDefaultsLayer) Apply(ctx context.Context, rctx ResolveContext, cfg EffectiveConfig) (EffectiveConfig, error) {
	tpl := rctx.Producer.Template
	if tpl == nil {
		return cfg, nil
	}

	cfg.Meta.TemplateCode = tpl.Code

	// Tokens: usar defaults de la template.
	cfg.Tokens = EffectiveTokens{
		Colors:  tpl.DefaultColors,
		Fonts:   tpl.DefaultFonts,
		Radius:  tpl.DefaultRadius,
		Shadows: tpl.DefaultShadows,
	}

	// Módulos por defecto de la template — todos habilitados.
	tmods, err := l.templatesRepo.GetTemplateModulesRepository(tpl.ID)
	if err != nil {
		return cfg, err
	}
	modules := make([]EffectiveModule, 0, len(tmods))
	for _, tm := range tmods {
		modules = append(modules, EffectiveModule{
			ID:      tm.Module.ID,
			Code:    tm.Module.Code,
			Name:    tm.Module.Name,
			IsCore:  tm.Module.IsCore,
			Enabled: true,
			Source:  l.Name(),
		})
	}
	cfg.Modules = modules

	// Páginas por defecto de la template.
	tpages, err := l.templatesRepo.GetTemplatePagesRepository(tpl.ID)
	if err != nil {
		return cfg, err
	}
	for _, p := range tpages {
		if len(p.PuckJSONDefault) > 0 {
			cfg.Pages[p.PageType] = p.PuckJSONDefault
		}
	}

	return cfg, nil
}
