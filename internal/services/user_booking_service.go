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
}

func NewUserBookingService(bookings repositories.BookingRepository) UserBookingService {
	return &userBookingService{bookings: bookings}
}

func (s *userBookingService) Create(ctx context.Context, userID int64, request dto.CreateBookingRequest) (*models.Booking, error) {
	code, err := newBookingCode()
	if err != nil {
		return nil, err
	}
	return s.bookings.Create(ctx, repositories.CreateBookingInput{
		UserID:     userID,
		ShowtimeID: request.ShowtimeID,
		SeatIDs:    request.SeatIDs,
		Code:       code,
	})
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
