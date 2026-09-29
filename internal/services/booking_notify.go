package services

import (
	"context"
	"log/slog"
	"time"

	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
)

const (
	notifyTimeout      = 15 * time.Second
	notifyMaxAttempts  = 3
	notifyRetryBackoff = 2 * time.Second
)

type BookingMailer interface {
	SendBookingCreated(ctx context.Context, booking *models.Booking) error
	SendBookingPaid(ctx context.Context, booking *models.Booking) error
}

func notifyBooking(bookings repositories.BookingRepository, id int64, event string, sends ...func(context.Context, *models.Booking) error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("notify booking panicked", "event", event, "booking_id", id, "panic", r)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), notifyTimeout)
	booking, err := bookings.FindDetail(ctx, id)
	cancel()
	if err != nil {
		slog.Error("load booking for notifications", "event", event, "booking_id", id, "error", err)
		return
	}
	for _, send := range sends {
		if err := sendWithRetry(booking, send); err != nil {
			slog.Error("notify booking", "event", event, "code", booking.Code, "user_id", booking.UserID, "attempts", notifyMaxAttempts, "error", err)
		}
	}
}

// sendWithRetry gives every attempt its own timeout and waits 2s, 4s, ... between attempts.
func sendWithRetry(booking *models.Booking, send func(context.Context, *models.Booking) error) error {
	var err error
	for attempt := 1; attempt <= notifyMaxAttempts; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), notifyTimeout)
		err = send(ctx, booking)
		cancel()
		if err == nil {
			return nil
		}
		if attempt < notifyMaxAttempts {
			slog.Warn("notify booking attempt failed", "code", booking.Code, "attempt", attempt, "error", err)
			time.Sleep(notifyRetryBackoff << (attempt - 1))
		}
	}
	return err
}
