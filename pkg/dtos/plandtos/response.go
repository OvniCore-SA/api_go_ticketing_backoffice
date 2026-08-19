package plandtos

import (
	"time"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
)

type ResponsePlan struct {
	ID          uint      `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (r *ResponsePlan) FromEntity(e entities.Plan) {
	r.ID = e.ID
	r.Code = e.Code
	r.Name = e.Name
	r.Description = e.Description
	r.IsActive = e.IsActive
	r.CreatedAt = e.CreatedAt
	r.UpdatedAt = e.UpdatedAt
}

type ResponsePlans struct {
	Plans []ResponsePlan `json:"plans"`
}

func (r *ResponsePlans) FromEntities(list []entities.Plan) {
	r.Plans = make([]ResponsePlan, len(list))
	for i, e := range list {
		r.Plans[i].FromEntity(e)
	}
}
