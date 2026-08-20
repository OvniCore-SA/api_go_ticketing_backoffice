package store

import (
	"context"
	"strings"

	pcfg "github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producer_config"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/domains"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/producers"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/storedtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"github.com/gofiber/fiber/v2"
)

type Service interface {
	GetTenantByHostService(ctx context.Context, rawHost string) (storedtos.ResponseTenantByHost, error)
}

type service struct {
	producersRepo producers.Repository
	domainsRepo   domains.Repository
	resolver      *pcfg.Resolver
}

func NewStoreService(
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

func cleanHost(raw string) string {
	host := strings.TrimSpace(raw)
	host = strings.ToLower(host)
	if idx := strings.Index(host, "://"); idx != -1 {
		host = host[idx+3:]
	}
	if idx := strings.Index(host, "/"); idx != -1 {
		host = host[:idx]
	}
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}
	return host
}

func (s *service) GetTenantByHostService(ctx context.Context, rawHost string) (storedtos.ResponseTenantByHost, error) {
	host := cleanHost(rawHost)
	if host == "" {
		return storedtos.ResponseTenantByHost{}, fiber.NewError(fiber.StatusBadRequest, storedtos.ErrHostRequired.Error())
	}

	var producer entities.Producer
	var domainMeta storedtos.DomainMeta
	foundDomain := false

	// 1. Buscar coincidencia en la tabla domains
	d, err := s.domainsRepo.GetOneRepository(filters.DomainFilter{Domain: host})
	if err == nil {
		p, errP := s.producersRepo.GetProducerRepository(filters.ProducerFilter{ID: d.ProducerID})
		if errP == nil {
			producer = p
			foundDomain = true
			domainMeta = storedtos.DomainMeta{
				Domain:   d.Domain,
				Type:     d.Type,
				Verified: d.VerifiedAt != nil,
			}
		}
	}

	// 2. Si no se encontró en la tabla domains, buscar en producers por slug / subdominio
	if !foundDomain {
		p, errP := s.producersRepo.GetProducerRepository(filters.ProducerFilter{Slug: host})
		if errP == nil {
			producer = p
			foundDomain = true
			domainMeta = storedtos.DomainMeta{
				Domain:   host,
				Type:     entities.DomainTypeSubdomain,
				Verified: true,
			}
		} else if strings.Contains(host, ".") {
			parts := strings.Split(host, ".")
			if len(parts) > 0 && parts[0] != "" {
				pSub, errSub := s.producersRepo.GetProducerRepository(filters.ProducerFilter{Slug: parts[0]})
				if errSub == nil {
					producer = pSub
					foundDomain = true
					domainMeta = storedtos.DomainMeta{
						Domain:   host,
						Type:     entities.DomainTypeSubdomain,
						Verified: true,
					}
				}
			}
		}
	}

	if !foundDomain {
		return storedtos.ResponseTenantByHost{}, fiber.NewError(fiber.StatusNotFound, storedtos.ErrTenantNotFound.Error())
	}

	// 3. Resolver la configuración efectiva usando Resolver
	cfg, err := s.resolver.Resolve(ctx, producer.ID)
	if err != nil {
		return storedtos.ResponseTenantByHost{}, fiber.NewError(fiber.StatusInternalServerError, "error al resolver la configuración del tenant")
	}

	// 4. Filtrar módulos activos (enabled = true)
	activeModules := make([]storedtos.ActiveModule, 0, len(cfg.Modules))
	for _, m := range cfg.Modules {
		if !m.Enabled {
			continue
		}
		activeModules = append(activeModules, storedtos.ActiveModule{
			ID:     m.ID,
			Code:   m.Code,
			Name:   m.Name,
			IsCore: m.IsCore,
		})
	}

	// 5. Mapear variantes de componentes
	variants := make(map[uint]storedtos.ActiveVariant, len(cfg.Variants))
	for k, v := range cfg.Variants {
		variants[k] = storedtos.ActiveVariant{
			ModuleID:           v.ModuleID,
			ComponentVariantID: v.ComponentVariantID,
			VariantCode:        v.VariantCode,
		}
	}

	return storedtos.ResponseTenantByHost{
		Tenant: storedtos.TenantMeta{
			ID:           producer.ID,
			Name:         producer.Name,
			Slug:         producer.Slug,
			ContactEmail: producer.ContactEmail,
			Status:       producer.Status,
			Suspended:    producer.Status == entities.ProducerStatusSuspended,
			TemplateID:   producer.TemplateID,
			PlanID:       producer.PlanID,
		},
		Domain:        domainMeta,
		ActiveModules: activeModules,
		Modules:       activeModules,
		Variants:      variants,
		Tokens: storedtos.ActiveTokens{
			Colors:  cfg.Tokens.Colors,
			Fonts:   cfg.Tokens.Fonts,
			Radius:  cfg.Tokens.Radius,
			Shadows: cfg.Tokens.Shadows,
		},
		Pages: cfg.Pages,
	}, nil
}
