package services

import (
	"context"
	"log/slog"
	"time"

	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
)

const notifyTimeout = 15 * time.Second

type BookingMailer interface {
	SendBookingCreated(ctx context.Context, booking *models.Booking) error
	SendBookingPaid(ctx context.Context, booking *models.Booking) error
}

func notifyBooking(bookings repositories.BookingRepository, id int64, event string, send func(context.Context, *models.Booking) error) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("notify booking panicked", "event", event, "booking_id", id, "panic", r)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), notifyTimeout)
	defer cancel()
	booking, err := bookings.FindDetail(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "load booking for notifications", "event", event, "booking_id", id, "error", err)
		return
	}
	if err := send(ctx, booking); err != nil {
		slog.ErrorContext(ctx, "send booking email", "event", event, "code", booking.Code, "to", booking.User.Email, "error", err)
	}
}
