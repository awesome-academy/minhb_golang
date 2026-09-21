package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"cinema-booking/config"
	"cinema-booking/internal/jobs"
	"cinema-booking/internal/repositories"
	"cinema-booking/pkg/db"
	"cinema-booking/pkg/logger"
)

func main() {
	if err := run(); err != nil {
		slog.Error("worker stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if _, err := logger.New(cfg.LogLevel, cfg.LogFormat); err != nil {
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("worker started", "hold_expiry_interval", cfg.HoldExpiryInterval)
	jobs.NewHoldExpiryJob(repositories.NewBookingRepository(database), cfg.HoldExpiryInterval).Run(ctx)
	slog.Info("worker stopped")
	return nil
}
