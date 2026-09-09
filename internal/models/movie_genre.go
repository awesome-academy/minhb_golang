package models

import "time"

type MovieGenre struct {
	MovieID   int64 `gorm:"primaryKey"`
	GenreID   int64 `gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
