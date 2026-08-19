package auth

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/authdtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
)

// IAuthService define las operaciones de autenticación del backoffice.
// Este servicio SOLO expone flujos internos de SuperAdmin: login, refresh,
// obtención del perfil actual y cambio de contraseña. No hay registro público,
// verificación de email ni recuperación pública de contraseña.
type IAuthService interface {
	// Público.
	Login(dto authdtos.LoginRequest) (authdtos.LoginResponse, error)

	// Protegidos — el userID proviene del JWT validado por ValidateToken.
	RefreshService(userID uint) (authdtos.LoginResponse, error)
	GetUserService(filter filters.UserFilter) (authdtos.ResponseUser, error)
	ChangePassword(request authdtos.RequestChangePassword) error

	// Interno — usado por el middleware y para emitir tokens tras seedear un
	// SuperAdmin inicial desde el bootstrap.
	GetTokensService(user entities.User) (authdtos.LoginResponse, error)
}
