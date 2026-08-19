package producerdtos

import (
	"regexp"
	"strings"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/commons"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
)

// slugPattern permite: minúsculas, dígitos y guion medio. Debe empezar y
// terminar con letra/dígito. Máximo 63 caracteres (límite DNS).
// Cumple RFC 1035 para que el slug sea usable como subdominio.
var slugPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

type RequestCreateProducer struct {
	Name         string `json:"name"`
	Slug         string `json:"slug"`
	ContactEmail string `json:"contact_email"`
	Status       string `json:"status"`
	TemplateID   *uint  `json:"template_id"`
	PlanID       *uint  `json:"plan_id"`
}

func (r *RequestCreateProducer) Validate() error {
	if commons.StringIsEmpty(r.Name) {
		return ErrNameRequired
	}
	if commons.StringIsEmpty(r.Slug) {
		return ErrSlugRequired
	}
	// Normalizar antes de validar formato.
	r.Slug = strings.ToLower(strings.TrimSpace(r.Slug))
	if !slugPattern.MatchString(r.Slug) {
		return ErrSlugInvalid
	}
	if r.ContactEmail != "" && !commons.IsEmailValid(r.ContactEmail) {
		return ErrEmailInvalid
	}
	if r.Status != "" && !isValidStatus(r.Status) {
		return ErrInvalidStatus
	}
	return nil
}

func (r *RequestCreateProducer) ToEntity() entities.Producer {
	status := entities.ProducerStatusActive
	if r.Status != "" {
		status = r.Status
	}
	return entities.Producer{
		Name:         r.Name,
		Slug:         r.Slug,
		ContactEmail: r.ContactEmail,
		Status:       status,
		TemplateID:   r.TemplateID,
		PlanID:       r.PlanID,
	}
}

// RequestUpdateProducer intencionalmente NO expone `slug` — es la clave
// pública del tenant (aparece en el subdominio y en URLs). Cambiarla
// después del alta rompería enlaces, cachés y DNS. Si un cliente quiere
// cambiar de identidad pública, se crea un Producer nuevo.
// Tampoco expone `template_id` — la template define el seed inicial y
// no tiene sentido semántico "cambiar la template" a un tenant ya
// configurado (para eso está el futuro apply-template).
type RequestUpdateProducer struct {
	Name         string `json:"name"`
	ContactEmail string `json:"contact_email"`
	Status       string `json:"status"`
	PlanID       *uint  `json:"plan_id"`
}

func (r *RequestUpdateProducer) Validate() error {
	if r.ContactEmail != "" && !commons.IsEmailValid(r.ContactEmail) {
		return ErrEmailInvalid
	}
	if r.Status != "" && !isValidStatus(r.Status) {
		return ErrInvalidStatus
	}
	return nil
}

type RequestListProducers struct {
	Search string `query:"search"`
	Status string `query:"status"`
	PlanID uint   `query:"plan_id"`
	Page   int    `query:"page"`
	Limit  int    `query:"limit"`
}

func isValidStatus(s string) bool {
	return s == entities.ProducerStatusActive || s == entities.ProducerStatusSuspended
}
