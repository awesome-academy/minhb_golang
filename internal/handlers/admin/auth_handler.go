package admin

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	"cinema-booking/internal/dto"
	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/middleware"
	"cinema-booking/internal/services"
	"cinema-booking/internal/utils"
)

type LoginView struct {
	Title     string
	CSRFToken string
	Email     string
	Error     string
}

type AuthHandler struct {
	authService services.AdminAuthService
	cookie      middleware.AdminSessionCookie
}

func NewAuthHandler(authService services.AdminAuthService, cookie middleware.AdminSessionCookie) *AuthHandler {
	return &AuthHandler{authService: authService, cookie: cookie}
}

func (h *AuthHandler) LoginPage(c *echo.Context) error {
	return c.Render(http.StatusOK, "admin/login", LoginView{
		Title:     "Sign in",
		CSRFToken: csrfToken(c),
	})
}

func (h *AuthHandler) Login(c *echo.Context) error {
	var request dto.AdminLoginRequest
	if err := utils.BindAndValidate(c, &request); err != nil {
		return h.renderLoginError(c, c.FormValue("email"))
	}

	sessionID, err := h.authService.Login(c.Request().Context(), request.Email, request.Password)
	if errors.Is(err, apperrors.ErrInvalidCredentials) {
		return h.renderLoginError(c, request.Email)
	}
	if err != nil {
		return utils.ServiceError(err)
	}

	h.cookie.Set(c, sessionID)
	return c.Redirect(http.StatusSeeOther, middleware.AdminPathPrefix)
}

func (h *AuthHandler) Logout(c *echo.Context) error {
	if sessionCookie, err := c.Cookie(middleware.AdminSessionCookieName); err == nil && sessionCookie.Value != "" {
		if err := h.authService.Logout(c.Request().Context(), sessionCookie.Value); err != nil {
			c.Logger().Error("failed to delete admin session", "error", err)
		}
	}
	h.cookie.Clear(c)
	return c.Redirect(http.StatusSeeOther, middleware.AdminLoginPath)
}

func (h *AuthHandler) renderLoginError(c *echo.Context, email string) error {
	return c.Render(http.StatusUnprocessableEntity, "admin/login", LoginView{
		Title:     "Sign in",
		CSRFToken: csrfToken(c),
		Email:     email,
		Error:     "Invalid email or password",
	})
}

func csrfToken(c *echo.Context) string {
	token, _ := c.Get(middleware.CSRFContextKey).(string)
	return token
}
