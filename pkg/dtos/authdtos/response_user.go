package authdtos

import "github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"

type ResponseUser struct {
	ID       uint   `json:"id"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Active   bool   `json:"active"`
	RoleID   uint   `json:"role_id"`
	RoleCode string `json:"role_code"`
}

func (r *ResponseUser) FromEntity(entity entities.User) {
	r.ID = entity.ID
	r.Name = entity.Name
	r.Email = entity.Email
	r.Active = entity.Active
	r.RoleID = entity.RoleID
	r.RoleCode = entity.Role.Code
}
