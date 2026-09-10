package middleware

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"
	"gorm.io/gorm"

	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
)

const (
	AdminUserContextKey = "admin_user"
	AdminLoginPath      = AdminPathPrefix + "/login"
)

func RequireAdminSession(cookie AdminSessionCookie, sessions repositories.AdminSessionRepository, users repositories.UserRepository) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			sessionCookie, err := c.Cookie(AdminSessionCookieName)
			if err != nil || sessionCookie.Value == "" {
				return redirectToLogin(c, cookie)
			}

			ctx := c.Request().Context()
			userID, err := sessions.FindUserID(ctx, sessionCookie.Value)
			if err != nil {
				if !errors.Is(err, apperrors.ErrSessionNotFound) {
					c.Logger().Error("admin session lookup failed", "error", err)
				}
				return redirectToLogin(c, cookie)
			}

			user, err := users.FindByID(ctx, userID)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				c.Logger().Error("admin user lookup failed", "error", err)
				return redirectToLogin(c, cookie)
			}
			if err != nil || user.Role != models.UserRoleAdmin {
				if err := sessions.Delete(ctx, sessionCookie.Value); err != nil {
					c.Logger().Error("failed to delete stale admin session", "error", err)
				}
				return redirectToLogin(c, cookie)
			}

			c.Set(AdminUserContextKey, user)
			return next(c)
		}
	}
}

func AdminUser(c *echo.Context) *models.User {
	user, _ := c.Get(AdminUserContextKey).(*models.User)
	return user
}

func redirectToLogin(c *echo.Context, cookie AdminSessionCookie) error {
	cookie.Clear(c)
	return c.Redirect(http.StatusFound, AdminLoginPath)
}
