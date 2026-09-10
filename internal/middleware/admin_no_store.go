package middleware

import "github.com/labstack/echo/v5"

func AdminNoStore() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			if IsAdminPath(c.Request().URL.Path) {
				c.Response().Header().Set(echo.HeaderCacheControl, "no-store")
			}
			return next(c)
		}
	}
}
