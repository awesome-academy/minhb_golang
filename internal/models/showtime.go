package models

import "time"

type ShowtimeFormat string

const (
	ShowtimeFormat2D   ShowtimeFormat = "2D"
	ShowtimeFormat3D   ShowtimeFormat = "3D"
	ShowtimeFormatIMAX ShowtimeFormat = "IMAX"
)

type ShowtimeStatus string

const (
	ShowtimeStatusScheduled ShowtimeStatus = "scheduled"
	ShowtimeStatusCancelled ShowtimeStatus = "cancelled"
)

type Showtime struct {
	ID        int64          `gorm:"primaryKey"`
	MovieID   int64          `gorm:"not null"`
	RoomID    int64          `gorm:"not null"`
	StartsAt  time.Time      `gorm:"not null"`
	EndsAt    time.Time      `gorm:"not null"`
	Format    ShowtimeFormat `gorm:"type:showtime_format;not null;default:'2D'"`
	Status    ShowtimeStatus `gorm:"type:showtime_status;not null;default:'scheduled'"`
	CreatedAt time.Time
	UpdatedAt time.Time

	Movie  Movie
	Room   Room
	Prices []ShowtimePrice
}
