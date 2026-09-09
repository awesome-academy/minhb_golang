package models

import (
	"time"

	"gorm.io/gorm"
)

type Room struct {
	ID        int64  `gorm:"primaryKey"`
	TheaterID int64  `gorm:"not null;uniqueIndex:rooms_theater_id_name_key"`
	Name      string `gorm:"not null;uniqueIndex:rooms_theater_id_name_key"`
	IsActive  bool   `gorm:"not null"`
	DeletedAt gorm.DeletedAt
	CreatedAt time.Time
	UpdatedAt time.Time

	Theater Theater
	Seats   []Seat
}
