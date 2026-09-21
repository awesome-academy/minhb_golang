package dto

import (
	"sort"
	"time"

	"cinema-booking/internal/models"
)

const (
	SeatStatusAvailable = "available"
	SeatStatusHeld      = "held"
	SeatStatusSold      = "sold"
	SeatStatusBlocked   = "blocked"
)

type ShowtimeDetail struct {
	ID       int64           `json:"id" example:"1"`
	StartsAt time.Time       `json:"startsAt" example:"2026-09-19T12:00:00Z"`
	EndsAt   time.Time       `json:"endsAt" example:"2026-09-19T14:30:00Z"`
	Format   string          `json:"format" example:"2D"`
	Movie    MovieSummary    `json:"movie"`
	Theater  TheaterResponse `json:"theater"`
	Room     RoomResponse    `json:"room"`
}

type SeatPrice struct {
	SeatTypeID int64  `json:"seatTypeId" example:"1"`
	Code       string `json:"code" example:"standard"`
	Name       string `json:"name" example:"Standard"`
	Price      string `json:"price" example:"75000.00"`
}

type SeatResponse struct {
	ID       int64   `json:"id" example:"1"`
	Row      string  `json:"row" example:"A"`
	Number   int     `json:"number" example:"1"`
	SeatType string  `json:"seatType" example:"vip"`
	Price    *string `json:"price" example:"95000.00"`
	Status   string  `json:"status" example:"available" enums:"available,held,sold,blocked"`
}

type SeatMapResponse struct {
	Showtime ShowtimeDetail `json:"showtime"`
	Prices   []SeatPrice    `json:"prices"`
	Seats    []SeatResponse `json:"seats"`
}

func NewSeatMapResponse(showtime *models.Showtime, seats []models.Seat, statuses map[int64]models.TicketStatus) SeatMapResponse {
	prices := make([]SeatPrice, 0, len(showtime.Prices))
	priceByType := make(map[int64]string, len(showtime.Prices))
	for _, price := range showtime.Prices {
		formatted := price.Price.StringFixed(2)
		priceByType[price.SeatTypeID] = formatted
		prices = append(prices, SeatPrice{
			SeatTypeID: price.SeatTypeID,
			Code:       price.SeatType.Code,
			Name:       price.SeatType.Name,
			Price:      formatted,
		})
	}
	sort.Slice(prices, func(i, j int) bool { return prices[i].SeatTypeID < prices[j].SeatTypeID })

	items := make([]SeatResponse, 0, len(seats))
	for i := range seats {
		items = append(items, newSeatResponse(&seats[i], priceByType, statuses[seats[i].ID]))
	}
	return SeatMapResponse{Showtime: newShowtimeDetail(showtime), Prices: prices, Seats: items}
}

func newShowtimeDetail(showtime *models.Showtime) ShowtimeDetail {
	return ShowtimeDetail{
		ID:       showtime.ID,
		StartsAt: showtime.StartsAt.UTC(),
		EndsAt:   showtime.EndsAt.UTC(),
		Format:   string(showtime.Format),
		Movie:    NewMovieSummary(&showtime.Movie),
		Theater:  NewTheaterResponse(&showtime.Room.Theater),
		Room:     RoomResponse{ID: showtime.Room.ID, Name: showtime.Room.Name},
	}
}

func newSeatResponse(seat *models.Seat, priceByType map[int64]string, ticket models.TicketStatus) SeatResponse {
	response := SeatResponse{
		ID:       seat.ID,
		Row:      seat.RowLabel,
		Number:   int(seat.SeatNumber),
		SeatType: seat.SeatType.Code,
	}
	if price, ok := priceByType[seat.SeatTypeID]; ok {
		response.Price = &price
	}
	response.Status = seatStatus(seat, response.Price != nil, ticket)
	return response
}

func seatStatus(seat *models.Seat, priced bool, ticket models.TicketStatus) string {
	switch {
	case !seat.IsActive || !priced:
		return SeatStatusBlocked
	case ticket == models.TicketStatusPaid:
		return SeatStatusSold
	case ticket == models.TicketStatusHeld:
		return SeatStatusHeld
	default:
		return SeatStatusAvailable
	}
}
