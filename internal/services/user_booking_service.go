package services

import (
	"context"
	"crypto/rand"
	"log/slog"
	"time"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
)

const (
	bookingCodeAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	bookingCodeLength   = 8
	notifyTimeout       = 15 * time.Second
)

type UserBookingService interface {
	Create(ctx context.Context, userID int64, request dto.CreateBookingRequest) (*models.Booking, error)
}

type userBookingService struct {
	bookings      repositories.BookingRepository
	notifications AdminNotificationService
}

func NewUserBookingService(bookings repositories.BookingRepository, notifications AdminNotificationService) UserBookingService {
	return &userBookingService{bookings: bookings, notifications: notifications}
}

func (s *userBookingService) Create(ctx context.Context, userID int64, request dto.CreateBookingRequest) (*models.Booking, error) {
	code, err := newBookingCode()
	if err != nil {
		return nil, err
	}
	booking, err := s.bookings.Create(ctx, repositories.CreateBookingInput{
		UserID:     userID,
		ShowtimeID: request.ShowtimeID,
		SeatIDs:    request.SeatIDs,
		Code:       code,
	})
	if err != nil {
		return nil, err
	}
	go s.notifyCreated(booking.ID)
	return booking, nil
}

func (s *userBookingService) notifyCreated(id int64) {
	defer func() {
		if r := recover(); r != nil {
			slog.Error("notify booking created panicked", "booking_id", id, "panic", r)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), notifyTimeout)
	defer cancel()
	booking, err := s.bookings.FindDetail(ctx, id)
	if err != nil {
		slog.ErrorContext(ctx, "load booking for notifications", "booking_id", id, "error", err)
		return
	}
	if err := s.notifications.BookingCreated(ctx, booking); err != nil {
		slog.ErrorContext(ctx, "notify admins", "code", booking.Code, "error", err)
	}
}

func newBookingCode() (string, error) {
	buf := make([]byte, bookingCodeLength)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	code := make([]byte, bookingCodeLength)
	for i, b := range buf {
		code[i] = bookingCodeAlphabet[int(b)%len(bookingCodeAlphabet)]
	}
	return string(code), nil
}
