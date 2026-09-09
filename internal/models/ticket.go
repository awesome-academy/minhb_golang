package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type TicketStatus string

const (
	TicketStatusHeld     TicketStatus = "held"
	TicketStatusPaid     TicketStatus = "paid"
	TicketStatusReleased TicketStatus = "released"
)

type Ticket struct {
	ID          int64           `gorm:"primaryKey"`
	BookingID   int64           `gorm:"not null"`
	ShowtimeID  int64           `gorm:"not null"`
	SeatID      int64           `gorm:"not null"`
	Price       decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	Status      TicketStatus    `gorm:"type:ticket_status;not null;default:'held'"`
	QRCode      string          `gorm:"type:uuid;not null;uniqueIndex;default:gen_random_uuid()"`
	CheckedInAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Seat Seat
}
