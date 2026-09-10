package handlers

import (
	"github.com/labstack/echo/v5"

	"cinema-booking/internal/handlers/admin"
	"cinema-booking/internal/middleware"
	"cinema-booking/internal/services"
)

func RegisterRoutes(e *echo.Echo, healthService services.HealthService) {
	api := e.Group("/api")
	registerHealthRoutes(api, NewHealthHandler(healthService))

	adminGroup := e.Group(middleware.AdminPathPrefix)
	registerAdminRoutes(adminGroup, admin.NewAuthHandler())
}

func registerHealthRoutes(api *echo.Group, handler *HealthHandler) {
	api.GET("/health", handler.Check)
}

func registerAdminRoutes(g *echo.Group, auth *admin.AuthHandler) {
	g.GET("/login", auth.LoginPage)
}
