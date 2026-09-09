package models

import "time"

type Genre struct {
	ID        int64  `gorm:"primaryKey"`
	Name      string `gorm:"not null;uniqueIndex"`
	Slug      string `gorm:"not null;uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
