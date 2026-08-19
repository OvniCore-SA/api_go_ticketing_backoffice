package plandtos

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/commons"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
)

type RequestCreatePlan struct {
	Code        string `json:"code"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

func (r RequestCreatePlan) Validate() error {
	if commons.StringIsEmpty(r.Code) {
		return ErrPlanCodeRequired
	}
	if commons.StringIsEmpty(r.Name) {
		return ErrPlanNameRequired
	}
	return nil
}

func (r *RequestCreatePlan) ToEntity() entities.Plan {
	active := true
	if r.IsActive != nil {
		active = *r.IsActive
	}
	return entities.Plan{
		Code:        r.Code,
		Name:        r.Name,
		Description: r.Description,
		IsActive:    active,
	}
}

// RequestUpdatePlan intencionalmente NO expone `code` — es la clave lógica
// del plan (usada por producer.plan_id y por posibles referencias externas).
// Cambiarla rompería referencias silenciosamente. Si querés renombrar el
// plan cara al usuario, editá `name`.
type RequestUpdatePlan struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	IsActive    *bool  `json:"is_active"`
}

type RequestListPlans struct {
	Search   string `query:"search"`
	IsActive *bool  `query:"is_active"`
	Page     int    `query:"page"`
	Limit    int    `query:"limit"`
}
