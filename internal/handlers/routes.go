package handlers

import (
	"github.com/labstack/echo/v5"

	"cinema-booking/internal/services"
)

func RegisterRoutes(e *echo.Echo, healthService services.HealthService) {
	api := e.Group("/api")

	registerHealthRoutes(api, NewHealthHandler(healthService))
}

func registerHealthRoutes(api *echo.Group, handler *HealthHandler) {
	api.GET("/health", handler.Check)
}
