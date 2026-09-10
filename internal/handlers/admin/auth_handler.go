package admin

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"cinema-booking/internal/middleware"
)

type LoginView struct {
	Title     string
	CSRFToken string
	Email     string
	Error     string
}

type AuthHandler struct{}

func NewAuthHandler() *AuthHandler {
	return &AuthHandler{}
}

func (h *AuthHandler) LoginPage(c *echo.Context) error {
	token, _ := c.Get(middleware.CSRFContextKey).(string)
	return c.Render(http.StatusOK, "admin/login", LoginView{
		Title:     "Sign in",
		CSRFToken: token,
	})
}
