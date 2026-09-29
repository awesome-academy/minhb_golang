package jobs

import (
	"context"
	"log/slog"

	"cinema-booking/internal/repositories"
)

type HoldExpiryJob struct {
	bookings repositories.BookingRepository
}

func NewHoldExpiryJob(bookings repositories.BookingRepository) *HoldExpiryJob {
	return &HoldExpiryJob{bookings: bookings}
}

func (j *HoldExpiryJob) Run(ctx context.Context) {
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
