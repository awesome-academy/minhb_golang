package services

import (
	"context"
	"crypto/rand"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
)

const (
	bookingCodeAlphabet = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ"
	bookingCodeLength   = 8
)

type UserBookingService interface {
	Create(ctx context.Context, userID int64, request dto.CreateBookingRequest) (*models.Booking, error)
}

type userBookingService struct {
	bookings repositories.BookingRepository
	mailer   BookingMailer
}

func NewUserBookingService(bookings repositories.BookingRepository, mailer BookingMailer) UserBookingService {
	return &userBookingService{bookings: bookings, mailer: mailer}
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
	go notifyBooking(s.bookings, booking.ID, "created", s.mailer.SendBookingCreated)
	return booking, nil
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
