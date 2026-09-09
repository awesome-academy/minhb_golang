package main

import (
	"log/slog"
	"net/http"
	"os"

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
	"cinema-booking/web"
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
	if err := run(); err != nil {
		slog.Error("application stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	log, err := logger.New(cfg.LogLevel, cfg.LogFormat)
	if err != nil {
		return err
	}

	database, err := db.Connect(db.Options{
		DSN:          cfg.DatabaseURL,
		MaxOpenConns: cfg.DBMaxOpenConns,
		MaxIdleConns: cfg.DBMaxIdleConns,
		LogSQL:       cfg.DBLogSQL,
		Colorful:     cfg.DBLogColorful,
	})
	if err != nil {
		return err
	}
	defer func() { _ = db.Close(database) }()

	templates, err := web.ParseTemplates()
	if err != nil {
		return err
	}

	e := echo.New()
	e.Logger = log
	e.HTTPErrorHandler = handlers.HTTPErrorHandler
	e.Validator = utils.NewRequestValidator()
	e.Renderer = &echo.TemplateRenderer{Template: templates}
	e.StaticFS("/static", echo.MustSubFS(web.Files, "static"))

	e.Use(middleware.RequestLogger())
	e.Use(middleware.Recover())

	e.GET("/swaggers", func(c *echo.Context) error {
		return c.Redirect(http.StatusFound, "/swaggers/index.html")
	})
	e.GET("/swaggers/*", echoSwagger.WrapHandler)

	healthRepository := repositories.NewHealthRepository(database)
	healthService := services.NewHealthService(healthRepository)
	handlers.RegisterRoutes(e, healthService)

	return e.Start(cfg.HTTPAddr)
}
