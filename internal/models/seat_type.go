package models

import "time"

type SeatType struct {
	ID        int64  `gorm:"primaryKey"`
	Code      string `gorm:"not null;uniqueIndex"`
	Name      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
