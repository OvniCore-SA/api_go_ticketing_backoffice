package handlers

import (
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/domains/auth"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/dtos/authdtos"
	"github.com/OvniCore-SA/api_go_ticketing_backoffice/pkg/filters"
	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	service auth.IAuthService
}

func NewAuthHandler(service auth.IAuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Login — POST /auth/login (público).
func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req authdtos.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": auth.MsgInvalidParams})
	}

	resp, err := h.service.Login(req)
	if err != nil {
		return handleAuthError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"data":    resp,
		"message": auth.MsgLoginSuccess,
	})
}

// Refresh — POST /auth/refresh (protegido).
func (h *AuthHandler) Refresh(c *fiber.Ctx) error {
	caller := authUserFromCtx(c)
	resp, err := h.service.RefreshService(caller.ID)
	if err != nil {
		return handleAuthError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"data":    resp,
		"message": auth.MsgRefreshSuccess,
	})
}

// Me — GET /auth/me (protegido).
func (h *AuthHandler) Me(c *fiber.Ctx) error {
	caller := authUserFromCtx(c)
	user, err := h.service.GetUserService(filters.UserFilter{ID: caller.ID})
	if err != nil {
		return handleAuthError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"data":    user,
		"message": auth.MsgUserFetched,
	})
}

// ChangePassword — PUT /auth/change-password (protegido).
func (h *AuthHandler) ChangePassword(c *fiber.Ctx) error {
	var req authdtos.RequestChangePassword
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": false, "message": auth.MsgInvalidParams})
	}

	// El userID se toma del token, nunca del body.
	req.UserID = authUserFromCtx(c).ID

	if err := h.service.ChangePassword(req); err != nil {
		return handleAuthError(c, err)
	}
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": auth.MsgPasswordChanged,
	})
}

// Logout — POST /auth/logout (protegido).
// Sin lista de tokens revocados; el cliente descarta el JWT localmente.
// Se deja el endpoint para simetría con el frontend.
func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  true,
		"message": auth.MsgLogoutSuccess,
	})
}
