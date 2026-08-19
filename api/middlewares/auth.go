package middlewares

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/auth"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/entities"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

type MiddlewareManager struct {
	authService auth.IAuthService
}

func NewMiddlewareManager(authSvc auth.IAuthService) MiddlewareManager {
	return MiddlewareManager{authService: authSvc}
}

// ValidateToken valida el JWT Bearer del header Authorization y garantiza que
// el titular del token siga existiendo en DB con rol SuperAdmin activo.
//
// Garantías:
//  1. Solo se acepta HMAC — rechaza "alg:none" y variantes RSA.
//  2. El usuario se recarga en cada request filtrando por rol SuperAdmin;
//     si fue eliminado, desactivado o su rol cambió, el token queda inválido.
//  3. El usuario autenticado queda en c.Locals("user") como authdtos.ResponseUser.
func (m *MiddlewareManager) ValidateToken() fiber.Handler {
	return func(c *fiber.Ctx) error {
		bearer := c.Get("Authorization")
		if bearer == "" {
			return fiber.NewError(fiber.StatusUnauthorized, MsgTokenRequired)
		}

		parts := strings.Split(bearer, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			return fiber.NewError(fiber.StatusUnauthorized, "formato de token inválido — esperado: Bearer <token>")
		}

		token, err := jwt.Parse(parts[1], func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("algoritmo de firma no permitido: %v", t.Header["alg"])
			}
			return []byte(os.Getenv("JWT_SECRET_KEY")), nil
		})
		if err != nil || !token.Valid {
			return fiber.NewError(fiber.StatusUnauthorized, MsgTokenInvalidExpired)
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, MsgTokenInvalid)
		}

		subStr, ok := claims["sub"].(string)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, MsgTokenInvalid)
		}
		userID, err := strconv.ParseUint(subStr, 10, 64)
		if err != nil || userID == 0 {
			return fiber.NewError(fiber.StatusUnauthorized, MsgTokenInvalid)
		}

		// Recarga desde DB — el filtro RoleCode fuerza que el usuario siga
		// siendo SuperAdmin. Si dejó de serlo, no hay acceso.
		user, err := m.authService.GetUserService(filters.UserFilter{
			ID:       uint(userID),
			RoleCode: entities.RoleCodeSuperAdmin,
		})
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, MsgUnauthorized)
		}

		if !user.Active {
			return fiber.NewError(fiber.StatusForbidden, MsgUnauthorized)
		}

		c.Locals("user", user)
		return c.Next()
	}
}
