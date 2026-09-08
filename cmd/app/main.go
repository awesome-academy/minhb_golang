package main

import (
	"net/http"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	echoSwagger "github.com/swaggo/echo-swagger/v2"

	_ "cinema-booking/api/swagger"
	"cinema-booking/config"
	"cinema-booking/internal/handlers"
	"cinema-booking/internal/repositories"
	"cinema-booking/internal/services"
	"cinema-booking/internal/utils"
	"cinema-booking/pkg/db"
	"cinema-booking/pkg/logger"
)

// @title Cinema Booking API
// @version 1.0
// @description REST API cho web đặt vé rạp chiếu phim.
// @BasePath /api
// @schemes http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Nhập `Bearer {access token}`.
func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	log, err := logger.New(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		panic(err)
	}

	database, err := db.Connect(db.Options{
		DSN:          cfg.DatabaseURL,
		MaxOpenConns: cfg.DBMaxOpenConns,
		MaxIdleConns: cfg.DBMaxIdleConns,
	})
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.Logger = log
	e.HTTPErrorHandler = handlers.HTTPErrorHandler
	e.Validator = utils.NewRequestValidator()

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.GET("/swaggers", func(c *echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/swaggers/index.html")
	})
	e.GET("/swaggers/*", echoSwagger.WrapHandler)

	healthRepository := repositories.NewHealthRepository(database)
	healthService := services.NewHealthService(healthRepository)
	handlers.RegisterRoutes(e, healthService)

	if err := e.Start(cfg.HTTPAddr); err != nil {
		e.Logger.Error("failed to start server", "error", err)
	}
}
