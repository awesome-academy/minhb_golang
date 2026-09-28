package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/robfig/cron/v3"

	"cinema-booking/config"
	"cinema-booking/internal/jobs"
	"cinema-booking/internal/repositories"
	"cinema-booking/internal/utils"
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

	holdExpiryJob := jobs.NewHoldExpiryJob(repositories.NewBookingRepository(database))
	movieStatusJob := jobs.NewMovieStatusJob(repositories.NewMovieRepository(database))

	scheduler := cron.New(cron.WithLocation(utils.Location))
	if _, err := scheduler.AddFunc(cfg.HoldExpiryCron, func() { holdExpiryJob.Run(ctx) }); err != nil {
		return fmt.Errorf("HOLD_EXPIRY_CRON %q: %w", cfg.HoldExpiryCron, err)
	}
	if _, err := scheduler.AddFunc(cfg.MovieStatusCron, func() { movieStatusJob.Run(ctx) }); err != nil {
		return fmt.Errorf("MOVIE_STATUS_CRON %q: %w", cfg.MovieStatusCron, err)
	}

	slog.Info("worker started", "hold_expiry_cron", cfg.HoldExpiryCron, "movie_status_cron", cfg.MovieStatusCron)
	holdExpiryJob.Run(ctx)
	movieStatusJob.Run(ctx)
	scheduler.Start()

	<-ctx.Done()
	<-scheduler.Stop().Done()
	slog.Info("worker stopped")
	return nil
}
