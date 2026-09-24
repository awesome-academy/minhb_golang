package dto

import (
	"strconv"
	"time"

	"cinema-booking/internal/models"
	"cinema-booking/internal/utils"
)

const (
	notificationTypeBookingCreated = "booking_created"
	notificationTimeLayout         = "Mon 02 Jan 2006 15:04"
)

type AdminNotification struct {
	Type       string    `json:"type"`
	BookingID  int64     `json:"bookingId"`
	ShowtimeID int64     `json:"showtimeId"`
	UserEmail  string    `json:"userEmail"`
	MovieTitle string    `json:"movieTitle"`
	Theater    string    `json:"theater"`
	Room       string    `json:"room"`
	StartsAt   string    `json:"startsAt"`
	Seats      []string  `json:"seats"`
	Total      string    `json:"total"`
	CreatedAt  time.Time `json:"createdAt"`
}

func NewBookingNotification(booking *models.Booking) AdminNotification {
	seats := make([]string, 0, len(booking.Tickets))
	for _, ticket := range booking.Tickets {
		seats = append(seats, ticket.Seat.RowLabel+strconv.Itoa(int(ticket.Seat.SeatNumber)))
	}
	return AdminNotification{
		Type:       notificationTypeBookingCreated,
		BookingID:  booking.ID,
		ShowtimeID: booking.ShowtimeID,
		UserEmail:  booking.User.Email,
		MovieTitle: booking.Showtime.Movie.Title,
		Theater:    booking.Showtime.Room.Theater.Name,
		Room:       booking.Showtime.Room.Name,
		StartsAt:   utils.FormatVN(booking.Showtime.StartsAt, notificationTimeLayout),
		Seats:      seats,
		Total:      booking.Subtotal.Sub(booking.DiscountAmount).StringFixed(2),
		CreatedAt:  booking.CreatedAt.UTC(),
	}
}
