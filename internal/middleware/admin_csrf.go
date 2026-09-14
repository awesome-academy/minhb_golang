package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	echomw "github.com/labstack/echo/v5/middleware"
)

const (
	AdminPathPrefix = "/admin"
	CSRFContextKey  = "csrf"
	CSRFFormField   = "_csrf"
)

func IsAdminPath(path string) bool {
	return path == AdminPathPrefix || strings.HasPrefix(path, AdminPathPrefix+"/")
}

func AdminCSRF(secure bool) echo.MiddlewareFunc {
	return echomw.CSRFWithConfig(echomw.CSRFConfig{
		Skipper: func(c *echo.Context) bool {
			return !IsAdminPath(c.Request().URL.Path)
		},
		TokenLookup:    "form:" + CSRFFormField + ",header:" + echo.HeaderXCSRFToken,
		ContextKey:     CSRFContextKey,
		CookiePath:     AdminPathPrefix,
		CookieHTTPOnly: true,
		CookieSecure:   secure,
		CookieSameSite: http.SameSiteLaxMode,
	})
}
