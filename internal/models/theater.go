package models

import (
	"time"

	"gorm.io/gorm"
)

type Theater struct {
	ID          int64  `gorm:"primaryKey"`
	Slug        string `gorm:"not null;uniqueIndex"`
	Name        string `gorm:"not null"`
	Address     string `gorm:"not null"`
	City        string `gorm:"not null"`
	Phone       *string
	Description *string
	ImageURL    *string
	IsActive    bool `gorm:"not null"`
	DeletedAt   gorm.DeletedAt
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Rooms []Room
}
