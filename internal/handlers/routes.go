package handlers

import (
	"github.com/labstack/echo/v5"

	"cinema-booking/internal/handlers/admin"
	"cinema-booking/internal/handlers/user"
	"cinema-booking/internal/middleware"
	"cinema-booking/internal/services"
)

func RegisterRoutes(
	e *echo.Echo,
	healthService services.HealthService,
	userAuthService services.UserAuthService,
	adminAuthService services.AdminAuthService,
	adminMovieService services.AdminMovieService,
	adminTheaterService services.AdminTheaterService,
	adminRoomService services.AdminRoomService,
	adminSeatService services.AdminSeatService,
	adminShowtimeService services.AdminShowtimeService,
	adminCookie middleware.AdminSessionCookie,
	adminSession echo.MiddlewareFunc,
	userAuth echo.MiddlewareFunc,
) {
	api := e.Group("/api")
	registerHealthRoutes(api, NewHealthHandler(healthService))
	registerUserRoutes(api, user.NewAuthHandler(userAuthService), userAuth)

	adminGroup := e.Group(middleware.AdminPathPrefix)
	registerAdminRoutes(
		adminGroup,
		admin.NewAuthHandler(adminAuthService, adminCookie),
		admin.NewDashboardHandler(),
		admin.NewMovieHandler(adminMovieService, adminCookie.Secure),
		admin.NewTheaterHandler(adminTheaterService, adminCookie.Secure),
		admin.NewRoomHandler(adminRoomService, adminCookie.Secure),
		admin.NewSeatHandler(adminSeatService, adminCookie.Secure),
		admin.NewShowtimeHandler(adminShowtimeService, adminCookie.Secure),
		adminSession,
	)
}

func registerHealthRoutes(api *echo.Group, handler *HealthHandler) {
	api.GET("/health", handler.Check)
}

func registerUserRoutes(api *echo.Group, auth *user.AuthHandler, userAuth echo.MiddlewareFunc) {
	g := api.Group("/auth")
	g.POST("/register", auth.Register)
	g.POST("/login", auth.Login)
	g.POST("/logout", auth.Logout, userAuth)
	g.GET("/me", auth.Me, userAuth)
}

func registerAdminRoutes(
	g *echo.Group,
	auth *admin.AuthHandler,
	dashboard *admin.DashboardHandler,
	movies *admin.MovieHandler,
	theaters *admin.TheaterHandler,
	rooms *admin.RoomHandler,
	seats *admin.SeatHandler,
	showtimes *admin.ShowtimeHandler,
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
	protected.GET("/theaters/:id/rooms", rooms.List)
	protected.GET("/theaters/:id/rooms/new", rooms.New)
	protected.POST("/theaters/:id/rooms", rooms.Create)
	protected.GET("/rooms/:id/edit", rooms.Edit)
	protected.POST("/rooms/:id", rooms.Update)
	protected.POST("/rooms/:id/status", rooms.ChangeStatus)
	protected.GET("/rooms/:id/seats", seats.Map)
	protected.POST("/rooms/:id/seats/generate", seats.Generate)
	protected.POST("/rooms/:id/seats/types", seats.ChangeRowTypes)
	protected.POST("/rooms/:id/seats/:seatId/status", seats.ChangeStatus)
	protected.GET("/showtimes", showtimes.List)
	protected.GET("/showtimes/new", showtimes.New)
	protected.POST("/showtimes", showtimes.Create)
	protected.GET("/showtimes/:id/edit", showtimes.Edit)
	protected.POST("/showtimes/:id", showtimes.Update)
	protected.POST("/showtimes/:id/cancel", showtimes.Cancel)
	protected.POST("/showtimes/:id/publish", showtimes.ChangePublished)
}
