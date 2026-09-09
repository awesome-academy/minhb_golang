package models

import (
	"time"

	"github.com/shopspring/decimal"
)

type ShowtimePrice struct {
	ID         int64           `gorm:"primaryKey"`
	ShowtimeID int64           `gorm:"not null;uniqueIndex:showtime_prices_showtime_id_seat_type_id_key"`
	SeatTypeID int64           `gorm:"not null;uniqueIndex:showtime_prices_showtime_id_seat_type_id_key"`
	Price      decimal.Decimal `gorm:"type:numeric(12,2);not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time

	SeatType SeatType
}
