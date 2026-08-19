package producerdtos

import (
	"time"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
)

type ResponseProducer struct {
	ID           uint      `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	ContactEmail string    `json:"contact_email"`
	Status       string    `json:"status"`
	TemplateID   *uint     `json:"template_id"`
	PlanID       *uint     `json:"plan_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func (r *ResponseProducer) FromEntity(e entities.Producer) {
	r.ID = e.ID
	r.Name = e.Name
	r.Slug = e.Slug
	r.ContactEmail = e.ContactEmail
	r.Status = e.Status
	r.TemplateID = e.TemplateID
	r.PlanID = e.PlanID
	r.CreatedAt = e.CreatedAt
	r.UpdatedAt = e.UpdatedAt
}

type ResponseProducers struct {
	Producers []ResponseProducer `json:"producers"`
}

func (r *ResponseProducers) FromEntities(list []entities.Producer) {
	r.Producers = make([]ResponseProducer, len(list))
	for i, e := range list {
		r.Producers[i].FromEntity(e)
	}
}
