package handlers

import (
	"github.com/labstack/echo/v5"

	"cinema-booking/internal/handlers/admin"
	"cinema-booking/internal/middleware"
	"cinema-booking/internal/services"
)

func RegisterRoutes(
	e *echo.Echo,
	healthService services.HealthService,
	adminAuthService services.AdminAuthService,
	adminCookie middleware.AdminSessionCookie,
	adminSession echo.MiddlewareFunc,
) {
	api := e.Group("/api")
	registerHealthRoutes(api, NewHealthHandler(healthService))

	adminGroup := e.Group(middleware.AdminPathPrefix)
	registerAdminRoutes(adminGroup, admin.NewAuthHandler(adminAuthService, adminCookie), admin.NewDashboardHandler(), adminSession)
}

func registerHealthRoutes(api *echo.Group, handler *HealthHandler) {
	api.GET("/health", handler.Check)
}

func registerAdminRoutes(g *echo.Group, auth *admin.AuthHandler, dashboard *admin.DashboardHandler, adminSession echo.MiddlewareFunc) {
	g.GET("/login", auth.LoginPage)
	g.POST("/login", auth.Login)
	g.POST("/logout", auth.Logout)

	protected := g.Group("", adminSession)
	protected.GET("", dashboard.Index)
}
