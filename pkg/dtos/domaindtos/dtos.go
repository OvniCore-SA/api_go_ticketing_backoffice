package domaindtos

import (
	"errors"
	"strings"
	"time"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/commons"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
)

var (
	ErrDomainNotFound = errors.New("dominio no encontrado")
	ErrDomainExists   = errors.New("el dominio ya está registrado")
	ErrDomainReq      = errors.New("el dominio es requerido")
	ErrTypeInvalid    = errors.New("el tipo de dominio no es válido")
)

type RequestCreateDomain struct {
	Domain string `json:"domain"`
	Type   string `json:"type"`
}

func (r *RequestCreateDomain) Validate() error {
	if commons.StringIsEmpty(r.Domain) {
		return ErrDomainReq
	}
	r.Domain = strings.ToLower(strings.TrimSpace(r.Domain))
	if r.Type != entities.DomainTypeSubdomain && r.Type != entities.DomainTypeCustom {
		return ErrTypeInvalid
	}
	return nil
}

func (r *RequestCreateDomain) ToEntity(producerID uint) entities.Domain {
	return entities.Domain{
		ProducerID: producerID,
		Domain:     r.Domain,
		Type:       r.Type,
	}
}

type ResponseDomain struct {
	ID         uint       `json:"id"`
	ProducerID uint       `json:"producer_id"`
	Domain     string     `json:"domain"`
	Type       string     `json:"type"`
	VerifiedAt *time.Time `json:"verified_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func (r *ResponseDomain) FromEntity(e entities.Domain) {
	r.ID = e.ID
	r.ProducerID = e.ProducerID
	r.Domain = e.Domain
	r.Type = e.Type
	r.VerifiedAt = e.VerifiedAt
	r.CreatedAt = e.CreatedAt
}

type ResponseDomains struct {
	Domains []ResponseDomain `json:"domains"`
}

func (r *ResponseDomains) FromEntities(list []entities.Domain) {
	r.Domains = make([]ResponseDomain, len(list))
	for i, e := range list {
		r.Domains[i].FromEntity(e)
	}
}
