package admin

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) LoginPage(c *echo.Context) error {
	return c.Render(http.StatusOK, "admin/login", map[string]any{
		"Title": "Sign in",
	})
}
