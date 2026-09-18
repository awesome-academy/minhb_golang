package dto

import (
	"time"

	"cinema-booking/internal/models"
)

type CreateBookingRequest struct {
	ShowtimeID int64   `json:"showtimeId" validate:"required,gt=0" example:"1"`
	SeatIDs    []int64 `json:"seatIds" validate:"required,min=1,max=8,unique,dive,gt=0"`
}

type BookingTicketResponse struct {
	SeatID int64  `json:"seatId" example:"12"`
	Price  string `json:"price" example:"75000.00"`
}

type BookingResponse struct {
	ID        int64                   `json:"id" example:"1"`
	Code      string                  `json:"code" example:"BK7F3K9Q"`
	ExpiresAt time.Time               `json:"expiresAt" example:"2026-09-19T11:30:00Z"`
	Total     string                  `json:"total" example:"150000.00"`
	Tickets   []BookingTicketResponse `json:"tickets"`
}

func NewBookingResponse(booking *models.Booking) BookingResponse {
	tickets := make([]BookingTicketResponse, 0, len(booking.Tickets))
	for _, ticket := range booking.Tickets {
		tickets = append(tickets, BookingTicketResponse{SeatID: ticket.SeatID, Price: ticket.Price.StringFixed(2)})
	}
	var expiresAt time.Time
	if booking.ExpiresAt != nil {
		expiresAt = booking.ExpiresAt.UTC()
	}
	return BookingResponse{
		ID:        booking.ID,
		Code:      booking.Code,
		ExpiresAt: expiresAt,
		Total:     booking.Subtotal.Sub(booking.DiscountAmount).StringFixed(2),
		Tickets:   tickets,
	}
}
