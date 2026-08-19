// Package public_storefront expone la vista de LECTURA PÚBLICA de la
// configuración de un tenant para que el frontend (Next.js Starter Kit)
// pueda renderizar la tienda white-label sin autenticación.
//
// Estos endpoints son un BRIDGE temporal en este servicio hasta que
// ticketing-core esté operativo (spec §5.1 los define allá). Cuando se
// migren, el contrato de respuesta (publicstorefrontdtos) queda igual —
// el frontend no cambia.
package public_storefront

import (
	"context"
	"errors"

	pcfg "github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producer_config"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/domains"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producers"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/publicstorefrontdtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Service interface {
	// ResolveDomainService mapea un host (subdomain o custom) al producer.
	// Es lo que llama el middleware Edge de Next.js para poblar Redis.
	ResolveDomainService(ctx context.Context, domain string) (publicstorefrontdtos.ResponseResolveDomain, error)

	// GetConfigBySlugService devuelve el EffectiveConfig público completo
	// del tenant, filtrado (sin AppliedLayers, sin módulos disabled).
	GetConfigBySlugService(ctx context.Context, slug string) (publicstorefrontdtos.ResponsePublicConfig, error)

	// GetPageBySlugService devuelve solo la página solicitada + tokens +
	// módulos habilitados. Es la petición más caliente del sistema.
	GetPageBySlugService(ctx context.Context, slug, pageType string) (publicstorefrontdtos.ResponsePublicPage, error)
}

type service struct {
	producersRepo producers.Repository
	domainsRepo   domains.Repository
	resolver      *pcfg.Resolver
}

func NewPublicStorefrontService(
	producersRepo producers.Repository,
	domainsRepo domains.Repository,
	resolver *pcfg.Resolver,
) Service {
	return &service{
		producersRepo: producersRepo,
		domainsRepo:   domainsRepo,
		resolver:      resolver,
	}
}

func (s *service) ResolveDomainService(ctx context.Context, domain string) (publicstorefrontdtos.ResponseResolveDomain, error) {
	if domain == "" {
		return publicstorefrontdtos.ResponseResolveDomain{}, fiber.NewError(fiber.StatusBadRequest, "dominio requerido")
	}

	d, err := s.domainsRepo.GetOneRepository(filters.DomainFilter{Domain: domain})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return publicstorefrontdtos.ResponseResolveDomain{}, fiber.NewError(fiber.StatusNotFound, publicstorefrontdtos.ErrDomainNotFound.Error())
		}
		return publicstorefrontdtos.ResponseResolveDomain{}, fiber.NewError(fiber.StatusInternalServerError, "error al resolver dominio")
	}

	producer, err := s.producersRepo.GetProducerRepository(filters.ProducerFilter{ID: d.ProducerID})
	if err != nil {
		return publicstorefrontdtos.ResponseResolveDomain{}, fiber.NewError(fiber.StatusNotFound, publicstorefrontdtos.ErrProducerNotFound.Error())
	}

	return publicstorefrontdtos.ResponseResolveDomain{
		ProducerID: producer.ID,
		Slug:       producer.Slug,
		Name:       producer.Name,
		Status:     producer.Status,
		Suspended:  producer.Status == entities.ProducerStatusSuspended,
		DomainType: d.Type,
	}, nil
}

func (s *service) GetConfigBySlugService(ctx context.Context, slug string) (publicstorefrontdtos.ResponsePublicConfig, error) {
	if slug == "" {
		return publicstorefrontdtos.ResponsePublicConfig{}, fiber.NewError(fiber.StatusBadRequest, "slug requerido")
	}

	producer, err := s.producersRepo.GetProducerRepository(filters.ProducerFilter{Slug: slug})
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return publicstorefrontdtos.ResponsePublicConfig{}, fiber.NewError(fiber.StatusNotFound, publicstorefrontdtos.ErrProducerNotFound.Error())
		}
		return publicstorefrontdtos.ResponsePublicConfig{}, fiber.NewError(fiber.StatusInternalServerError, "error al obtener producer")
	}

	cfg, err := s.resolver.Resolve(ctx, producer.ID)
	if err != nil {
		// El Resolver ya devuelve fiber.Error con detalle interno; para
		// consumo público lo despersonalizamos.
		return publicstorefrontdtos.ResponsePublicConfig{}, fiber.NewError(fiber.StatusInternalServerError, "error al resolver configuración")
	}

	return toPublicConfig(cfg), nil
}

func (s *service) GetPageBySlugService(ctx context.Context, slug, pageType string) (publicstorefrontdtos.ResponsePublicPage, error) {
	if pageType == "" {
		return publicstorefrontdtos.ResponsePublicPage{}, fiber.NewError(fiber.StatusBadRequest, publicstorefrontdtos.ErrPageTypeInvalid.Error())
	}

	full, err := s.GetConfigBySlugService(ctx, slug)
	if err != nil {
		return publicstorefrontdtos.ResponsePublicPage{}, err
	}

	page, ok := full.Pages[pageType]
	if !ok || len(page) == 0 {
		return publicstorefrontdtos.ResponsePublicPage{}, fiber.NewError(fiber.StatusNotFound, publicstorefrontdtos.ErrPageNotFound.Error())
	}

	return publicstorefrontdtos.ResponsePublicPage{
		Producer: full.Producer,
		PageType: pageType,
		PuckJSON: page,
		Tokens:   full.Tokens,
		Modules:  full.Modules,
	}, nil
}

// toPublicConfig transforma EffectiveConfig (interno, con debug info) en
// ResponsePublicConfig (sanitizado). Filtra:
//   - Módulos con enabled=false (el frontend no debe saber que existen).
//   - AppliedLayers (implementación interna).
//   - Campo Source de cada módulo.
func toPublicConfig(cfg pcfg.EffectiveConfig) publicstorefrontdtos.ResponsePublicConfig {
	modules := make([]publicstorefrontdtos.PublicModule, 0, len(cfg.Modules))
	for _, m := range cfg.Modules {
		if !m.Enabled {
			continue
		}
		modules = append(modules, publicstorefrontdtos.PublicModule{
			ID:     m.ID,
			Code:   m.Code,
			Name:   m.Name,
			IsCore: m.IsCore,
		})
	}

	variants := make(map[uint]publicstorefrontdtos.PublicVariant, len(cfg.Variants))
	for k, v := range cfg.Variants {
		variants[k] = publicstorefrontdtos.PublicVariant{
			ModuleID:           v.ModuleID,
			ComponentVariantID: v.ComponentVariantID,
			VariantCode:        v.VariantCode,
		}
	}

	return publicstorefrontdtos.ResponsePublicConfig{
		Producer: publicstorefrontdtos.PublicProducerMeta{
			ID:        cfg.Meta.ProducerID,
			Slug:      cfg.Meta.ProducerSlug,
			Name:      cfg.Meta.ProducerName,
			Status:    cfg.Meta.Status,
			Suspended: cfg.Meta.Suspended,
		},
		Modules:  modules,
		Variants: variants,
		Tokens: publicstorefrontdtos.PublicTokens{
			Colors:  cfg.Tokens.Colors,
			Fonts:   cfg.Tokens.Fonts,
			Radius:  cfg.Tokens.Radius,
			Shadows: cfg.Tokens.Shadows,
		},
		Pages: cfg.Pages,
	}
}
