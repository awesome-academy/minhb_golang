package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusExpired   BookingStatus = "expired"
)

type Booking struct {
	ID             int64           `gorm:"primaryKey"`
	Code           string          `gorm:"not null;uniqueIndex"`
	UserID         int64           `gorm:"not null"`
	ShowtimeID     int64           `gorm:"not null"`
	Status         BookingStatus   `gorm:"type:booking_status;not null;default:'pending'"`
	Subtotal       decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	DiscountAmount decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	Currency       string          `gorm:"type:char(3);not null;default:'USD'"`
	ExpiresAt      *time.Time
	ConfirmedAt    *time.Time
	Note           *string
	CreatedAt      time.Time
	UpdatedAt      time.Time

	User     User
	Showtime Showtime
	Tickets  []Ticket
}
