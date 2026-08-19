package producer_config

import (
	"context"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/design_tokens"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/modules"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/page_templates"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producer_component_variants"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producer_modules"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"gorm.io/gorm"
)

// ProducerOverridesLayer aplica la configuración real del producer que el
// SuperAdmin editó a mano. Es la FUENTE ÚNICA DE VERDAD del sistema
// (según CLAUDE.md) y por eso REEMPLAZA lo que dejó la capa anterior en
// vez de mergear — si el SuperAdmin decidió deshabilitar Galería, no
// queremos que la template la vuelva a colar.
//
// Los tokens SÍ se mergean campo a campo: si el producer solo sobrescribió
// colores, se conservan las fuentes/radios/sombras de la template. Es lo
// que espera el frontend cuando el SuperAdmin "solo cambió el color primary".
type ProducerOverridesLayer struct {
	producerModulesRepo  producer_modules.Repository
	producerVariantsRepo producer_component_variants.Repository
	tokensRepo           design_tokens.Repository
	pagesRepo            page_templates.Repository
	modulesRepo          modules.Repository
}

func NewProducerOverridesLayer(
	pmRepo producer_modules.Repository,
	pvRepo producer_component_variants.Repository,
	tokensRepo design_tokens.Repository,
	pagesRepo page_templates.Repository,
	modulesRepo modules.Repository,
) *ProducerOverridesLayer {
	return &ProducerOverridesLayer{
		producerModulesRepo:  pmRepo,
		producerVariantsRepo: pvRepo,
		tokensRepo:           tokensRepo,
		pagesRepo:            pagesRepo,
		modulesRepo:          modulesRepo,
	}
}

func (l *ProducerOverridesLayer) Name() string { return "producer_overrides" }

func (l *ProducerOverridesLayer) Apply(ctx context.Context, rctx ResolveContext, cfg EffectiveConfig) (EffectiveConfig, error) {
	producerID := rctx.ProducerID

	// ─── Modules (reemplaza) ──────────────────────────────────────────────
	pmods, err := l.producerModulesRepo.GetByProducerRepository(producerID)
	if err != nil {
		return cfg, err
	}
	if len(pmods) > 0 {
		modulesList := make([]EffectiveModule, 0, len(pmods))
		for _, pm := range pmods {
			modulesList = append(modulesList, EffectiveModule{
				ID:      pm.ModuleID,
				Code:    pm.Module.Code,
				Name:    pm.Module.Name,
				IsCore:  pm.Module.IsCore,
				Enabled: pm.Enabled,
				Source:  l.Name(),
			})
		}
		cfg.Modules = modulesList
	}

	// ─── Component variants (upsert por module) ───────────────────────────
	pvars, err := l.producerVariantsRepo.GetByProducerRepository(producerID)
	if err != nil {
		return cfg, err
	}
	for _, pv := range pvars {
		cfg.Variants[pv.ModuleID] = EffectiveVariant{
			ModuleID:           pv.ModuleID,
			ComponentVariantID: pv.ComponentVariantID,
			VariantCode:        pv.ComponentVariant.Code,
		}
	}

	// ─── Design tokens (merge campo a campo) ──────────────────────────────
	tokens, err := l.tokensRepo.GetByProducerRepository(producerID)
	if err == nil {
		if len(tokens.Colors) > 0 {
			cfg.Tokens.Colors = tokens.Colors
		}
		if len(tokens.Fonts) > 0 {
			cfg.Tokens.Fonts = tokens.Fonts
		}
		if len(tokens.Radius) > 0 {
			cfg.Tokens.Radius = tokens.Radius
		}
		if len(tokens.Shadows) > 0 {
			cfg.Tokens.Shadows = tokens.Shadows
		}
	}

	// ─── Pages (upsert por pageType) ──────────────────────────────────────
	pages, err := l.pagesRepo.GetByProducerRepository(producerID)
	if err != nil {
		return cfg, err
	}
	for _, p := range pages {
		if len(p.PuckJSON) > 0 {
			cfg.Pages[p.PageType] = p.PuckJSON
		}
	}

	// ─── Nombres de módulos que la capa anterior no cargó ─────────────────
	// Si TemplateDefaultsLayer no corrió (producer sin template) pero
	// ProducerModule sí trae módulos, hay que enriquecer con datos del
	// catálogo global para que el frontend tenga Code/Name/IsCore.
	l.enrichModuleMetadata(&cfg)

	return cfg, nil
}

// enrichModuleMetadata rellena Code/Name/IsCore para EffectiveModule que
// hayan quedado con datos vacíos por venir directamente de ProducerModule
// sin preload. Idempotente.
func (l *ProducerOverridesLayer) enrichModuleMetadata(cfg *EffectiveConfig) {
	for i := range cfg.Modules {
		if cfg.Modules[i].Code != "" {
			continue
		}
		module, err := l.modulesRepo.GetModuleRepository(filters.ModuleFilter{ID: cfg.Modules[i].ID})
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				continue // módulo eliminado del catálogo — quedará "huérfano" pero visible
			}
			continue
		}
		cfg.Modules[i].Code = module.Code
		cfg.Modules[i].Name = module.Name
		cfg.Modules[i].IsCore = module.IsCore
	}
}

