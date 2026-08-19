package auth

import (
	"errors"
	"net/http"
)

// ─── Tipo de error ────────────────────────────────────────────────────────────

// AuthError transporta el código HTTP y el mensaje al handler.
type AuthError struct {
	Code    int
	Message string
}

func (e *AuthError) Error() string { return e.Message }

func errBadRequest(msg string) error { return &AuthError{Code: http.StatusBadRequest, Message: msg} }
func errUnauthorized(msg string) error {
	return &AuthError{Code: http.StatusUnauthorized, Message: msg}
}
func errForbidden(msg string) error { return &AuthError{Code: http.StatusForbidden, Message: msg} }
func errInternal(msg string) error {
	return &AuthError{Code: http.StatusInternalServerError, Message: msg}
}

// ─── Errores de dominio — sentinels para errors.Is ───────────────────────────

var (
	ErrInvalidCredentials = errors.New("El correo electrónico o la contraseña que ingresaste no son correctos. Por favor verificá tus datos e intentá de nuevo.")
	ErrAccountInactive    = errors.New("Tu cuenta está inactiva. Contactá con soporte para reactivarla.")
	ErrNotSuperAdmin      = errors.New("Esta cuenta no tiene acceso al backoffice.")
)

// Mensajes de respuesta exitosa.
const (
	MsgLoginSuccess    = "¡Bienvenido de nuevo! Ingresaste correctamente al backoffice."
	MsgRefreshSuccess  = "Sesión renovada con éxito."
	MsgPasswordChanged = "Contraseña actualizada con éxito."
	MsgUserFetched     = "Usuario obtenido con éxito."
	MsgLogoutSuccess   = "Sesión cerrada."
)

// Mensajes de error que se exponen al cliente.
const (
	MsgInternalError      = "Ocurrió un error inesperado. Por favor intentá nuevamente más tarde."
	MsgInvalidParams      = "Los datos enviados no son válidos. Por favor revisá la información e intentá de nuevo."
	MsgTokenGenerationErr = "No pudimos generar tu sesión. Por favor intentá nuevamente."
	MsgUserNotFound       = "No encontramos tu cuenta."
	MsgPasswordUpdateErr  = "No pudimos actualizar tu contraseña. Por favor intentá nuevamente."
)
