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
	adminMovieService services.AdminMovieService,
	adminTheaterService services.AdminTheaterService,
	adminCookie middleware.AdminSessionCookie,
	adminSession echo.MiddlewareFunc,
) {
	api := e.Group("/api")
	registerHealthRoutes(api, NewHealthHandler(healthService))

	adminGroup := e.Group(middleware.AdminPathPrefix)
	registerAdminRoutes(
		adminGroup,
		admin.NewAuthHandler(adminAuthService, adminCookie),
		admin.NewDashboardHandler(),
		admin.NewMovieHandler(adminMovieService, adminCookie.Secure),
		admin.NewTheaterHandler(adminTheaterService, adminCookie.Secure),
		adminSession,
	)
}

func registerHealthRoutes(api *echo.Group, handler *HealthHandler) {
	api.GET("/health", handler.Check)
}

func registerAdminRoutes(
	g *echo.Group,
	auth *admin.AuthHandler,
	dashboard *admin.DashboardHandler,
	movies *admin.MovieHandler,
	theaters *admin.TheaterHandler,
	adminSession echo.MiddlewareFunc,
) {
	g.GET("/login", auth.LoginPage)
	g.POST("/login", auth.Login)
	g.POST("/logout", auth.Logout)

	protected := g.Group("", adminSession)
	protected.GET("", dashboard.Index)
	protected.GET("/movies", movies.List)
	protected.GET("/movies/new", movies.New)
	protected.POST("/movies", movies.Create)
	protected.GET("/movies/:id/edit", movies.Edit)
	protected.POST("/movies/:id", movies.Update)
	protected.POST("/movies/:id/delete", movies.Delete)
	protected.GET("/theaters", theaters.List)
	protected.GET("/theaters/new", theaters.New)
	protected.POST("/theaters", theaters.Create)
	protected.GET("/theaters/:id/edit", theaters.Edit)
	protected.POST("/theaters/:id", theaters.Update)
	protected.POST("/theaters/:id/status", theaters.ChangeStatus)
}
