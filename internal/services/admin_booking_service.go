package services

import (
	"context"
	"strings"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
)

type CounterSeatMap struct {
	Showtime *models.Showtime
	Seats    []models.Seat
	Bookings []models.Booking
}

type AdminBookingService interface {
	SeatMap(ctx context.Context, showtimeID int64) (*CounterSeatMap, error)
	Sell(ctx context.Context, showtimeID, adminID int64, form dto.AdminCounterSaleForm) (*models.Booking, error)
	Confirm(ctx context.Context, bookingID int64, form dto.AdminConfirmBookingForm) (*models.Booking, error)
}

type adminBookingService struct {
	showtimes repositories.ShowtimeRepository
	seats     repositories.SeatRepository
	bookings  repositories.BookingRepository
}

func NewAdminBookingService(showtimes repositories.ShowtimeRepository, seats repositories.SeatRepository, bookings repositories.BookingRepository) AdminBookingService {
	return &adminBookingService{showtimes: showtimes, seats: seats, bookings: bookings}
}

func (s *adminBookingService) SeatMap(ctx context.Context, showtimeID int64) (*CounterSeatMap, error) {
	showtime, err := s.showtimes.FindByID(ctx, showtimeID)
	if err != nil {
		return nil, err
	}
	seats, err := s.seats.ListByRoom(ctx, showtime.RoomID)
	if err != nil {
		return nil, err
	}
	bookings, err := s.bookings.ActiveByShowtime(ctx, showtimeID)
	if err != nil {
		return nil, err
	}
	return &CounterSeatMap{Showtime: showtime, Seats: seats, Bookings: bookings}, nil
}

func (s *adminBookingService) Sell(ctx context.Context, showtimeID, adminID int64, form dto.AdminCounterSaleForm) (*models.Booking, error) {
	code, err := newBookingCode()
	if err != nil {
		return nil, err
	}
	return s.bookings.CreateCounterSale(ctx, repositories.CreateBookingInput{
		UserID:     adminID,
		ShowtimeID: showtimeID,
		SeatIDs:    form.SeatIDs,
		Code:       code,
	})
}

func (s *adminBookingService) Confirm(ctx context.Context, bookingID int64, form dto.AdminConfirmBookingForm) (*models.Booking, error) {
	return s.bookings.ConfirmPayment(ctx, bookingID, strings.ToUpper(strings.TrimSpace(form.Code)))
}
