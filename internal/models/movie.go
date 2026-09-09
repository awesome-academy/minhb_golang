package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type MovieStatus string

const (
	MovieStatusComingSoon MovieStatus = "coming_soon"
	MovieStatusNowShowing MovieStatus = "now_showing"
	MovieStatusEnded      MovieStatus = "ended"
)

type AgeRating string

const (
	AgeRatingP   AgeRating = "P"
	AgeRatingK   AgeRating = "K"
	AgeRatingT13 AgeRating = "T13"
	AgeRatingT16 AgeRating = "T16"
	AgeRatingT18 AgeRating = "T18"
	AgeRatingC   AgeRating = "C"
)

type CastMember struct {
	Name string `json:"name"`
	Role string `json:"role"`
}

type CastMembers []CastMember

func (c CastMembers) Value() (driver.Value, error) {
	if c == nil {
		return nil, nil
	}
	return json.Marshal(c)
}

func (c *CastMembers) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*c = nil
		return nil
	case []byte:
		return json.Unmarshal(v, c)
	case string:
		return json.Unmarshal([]byte(v), c)
	default:
		return fmt.Errorf("cannot scan %T into CastMembers", src)
	}
}

type Movie struct {
	ID            int64  `gorm:"primaryKey"`
	Slug          string `gorm:"not null;uniqueIndex"`
	Title         string `gorm:"not null"`
	OriginalTitle *string
	Description   *string
	DurationMin   int       `gorm:"not null"`
	AgeRating     AgeRating `gorm:"type:age_rating;not null"`
	Language      *string
	Director      *string
	CastMembers   CastMembers `gorm:"type:jsonb"`
	PosterURL     *string
	BackdropURL   *string
	TrailerURL    *string
	ReleaseDate   time.Time   `gorm:"type:date;not null"`
	Status        MovieStatus `gorm:"type:movie_status;not null;default:'coming_soon'"`
	DeletedAt     gorm.DeletedAt
	CreatedAt     time.Time
	UpdatedAt     time.Time

	Genres []Genre `gorm:"many2many:movie_genres"`
}
