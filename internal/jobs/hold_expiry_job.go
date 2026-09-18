package jobs

import (
	"context"
	"log/slog"
	"time"

	"cinema-booking/internal/repositories"
)

type HoldExpiryJob struct {
	bookings repositories.BookingRepository
	interval time.Duration
}

func NewHoldExpiryJob(bookings repositories.BookingRepository, interval time.Duration) *HoldExpiryJob {
	return &HoldExpiryJob{bookings: bookings, interval: interval}
}

func (j *HoldExpiryJob) Run(ctx context.Context) {
	j.runOnce(ctx)
	ticker := time.NewTicker(j.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			j.runOnce(ctx)
		}
	}
}

func (j *HoldExpiryJob) runOnce(ctx context.Context) {
	expired, err := j.bookings.ExpireHolds(ctx)
	if err != nil {
		if ctx.Err() == nil {
			slog.ErrorContext(ctx, "expire holds", "error", err)
		}
		return
	}
	if expired > 0 {
		slog.InfoContext(ctx, "expired holds", "bookings", expired)
	}
}
