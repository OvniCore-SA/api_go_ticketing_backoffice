package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/authdtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type authService struct {
	repository AuthRepository
}

func NewAuthService(repository AuthRepository) IAuthService {
	return &authService{repository: repository}
}

// Login autentica un SuperAdmin del backoffice. Rechaza cualquier cuenta
// cuyo rol no sea RoleCodeSuperAdmin.
func (s *authService) Login(request authdtos.LoginRequest) (authdtos.LoginResponse, error) {
	if err := request.Validate(); err != nil {
		return authdtos.LoginResponse{}, errBadRequest(err.Error())
	}

	user, err := s.repository.GetUserByEmail(request.Email)
	if err != nil {
		// No revelar si el email existe o no (mitigación de enumeración).
		return authdtos.LoginResponse{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		return authdtos.LoginResponse{}, ErrInvalidCredentials
	}

	if !user.Active {
		return authdtos.LoginResponse{}, ErrAccountInactive
	}

	if user.Role.Code != entities.RoleCodeSuperAdmin {
		return authdtos.LoginResponse{}, ErrNotSuperAdmin
	}

	response, err := s.GetTokensService(user)
	if err != nil {
		fmt.Printf("[login] error firmando token: usuario=%d err=%s\n", user.ID, err)
		return authdtos.LoginResponse{}, errInternal(MsgTokenGenerationErr)
	}
	return response, nil
}

// RefreshService reemite tokens para el usuario ya validado por ValidateToken.
func (s *authService) RefreshService(userID uint) (authdtos.LoginResponse, error) {
	if userID < 1 {
		return authdtos.LoginResponse{}, errBadRequest("usuario no especificado")
	}
	user, err := s.repository.GetUserByFilter(filters.UserFilter{ID: userID})
	if err != nil {
		return authdtos.LoginResponse{}, errUnauthorized(MsgUserNotFound)
	}
	return s.GetTokensService(user)
}

// GetUserService recupera el usuario segun filtros arbitrarios. Usado por
// ValidateToken (por ID) y por handlers de /me.
func (s *authService) GetUserService(filter filters.UserFilter) (authdtos.ResponseUser, error) {
	user, err := s.repository.GetUserByFilter(filter)
	if err != nil {
		return authdtos.ResponseUser{}, err
	}

	// Guardarraíl adicional: si el filtro no forzó el rol, verificamos igual
	// que la cuenta sea SuperAdmin — ValidateToken depende de esto.
	if filter.RoleCode == "" && user.Role.Code != entities.RoleCodeSuperAdmin {
		return authdtos.ResponseUser{}, errForbidden(ErrNotSuperAdmin.Error())
	}

	var response authdtos.ResponseUser
	response.FromEntity(user)
	return response, nil
}

func (s *authService) ChangePassword(request authdtos.RequestChangePassword) error {
	if err := request.Validate(); err != nil {
		return errBadRequest(err.Error())
	}

	user, err := s.repository.GetUserByFilter(filters.UserFilter{
		ID:           request.UserID,
		SelectFields: []string{"id", "password", "active", "role_id"},
	})
	if err != nil {
		return errUnauthorized(MsgUserNotFound)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(request.Password)); err != nil {
		return ErrInvalidCredentials
	}

	newPassword, err := bcrypt.GenerateFromPassword([]byte(request.NewPassword), 14)
	if err != nil {
		return errInternal(MsgPasswordUpdateErr)
	}

	if err := s.repository.UpdateUserDataRepository(user.ID, map[string]interface{}{
		"password": string(newPassword),
	}); err != nil {
		return errInternal(MsgPasswordUpdateErr)
	}
	return nil
}

// GetTokensService firma un par (access, refresh) HS256 para el usuario dado.
// El access token dura 48h; el refresh 5 días. El rol viaja en el claim para
// que el middleware pueda validar SuperAdmin sin releer DB en cada request si
// se optimiza más adelante.
func (s *authService) GetTokensService(user entities.User) (authdtos.LoginResponse, error) {
	var response authdtos.LoginResponse

	claims := jwt.MapClaims{
		"iss": "GoAccess-backoffice",
		"sub": fmt.Sprintf("%d", user.ID),
		"user": map[string]interface{}{
			"id":        user.ID,
			"email":     user.Email,
			"name":      user.Name,
			"role_id":   user.RoleID,
			"role_code": user.Role.Code,
		},
		"exp": time.Now().Add(48 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		return response, fmt.Errorf("no se pudo firmar el token")
	}

	refreshClaims := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "GoAccess-backoffice",
		Subject:   fmt.Sprintf("%d", user.ID),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(5 * 24 * time.Hour)),
	})
	refresh, err := refreshClaims.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		return response, fmt.Errorf("no se pudo generar el refresh token")
	}

	response.Token = signed
	response.RefreshToken = refresh
	response.User.FromEntity(user)
	return response, nil
}

var _ IAuthService = (*authService)(nil)
