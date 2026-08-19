package authdtos

import (
	"fmt"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/commons"
)

type RequestChangePassword struct {
	UserID            uint   `json:"user_id"`
	Password          string `json:"password"`
	NewPassword       string `json:"new_password"`
	RepeatNewPassword string `json:"repeat_new_password"`
}

func (r *RequestChangePassword) Validate() error {
	if r.UserID < 1 {
		return fmt.Errorf("invalid user")
	}

	if commons.StringIsEmpty(r.Password) {
		return fmt.Errorf("Ingresa tu contraseña actual")
	}

	if commons.ContainsSpaces(r.NewPassword) || commons.ContainsSpaces(r.RepeatNewPassword) {
		return fmt.Errorf("Las contraseñas no pueden contener espacios en blanco")
	}

	err := commons.ValidatePassword(r.NewPassword)
	if err != nil {
		return err
	}

	err = commons.ValidatePassword(r.RepeatNewPassword)
	if err != nil {
		return err
	}

	if r.NewPassword != r.RepeatNewPassword {
		return fmt.Errorf("Las contraseñas no coinciden")
	}
	return nil
}
