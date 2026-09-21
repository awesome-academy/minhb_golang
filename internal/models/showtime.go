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
	ID          int64          `gorm:"primaryKey"`
	MovieID     int64          `gorm:"not null"`
	RoomID      int64          `gorm:"not null"`
	StartsAt    time.Time      `gorm:"not null"`
	EndsAt      time.Time      `gorm:"not null"`
	Format      ShowtimeFormat `gorm:"type:showtime_format;not null;default:'2D'"`
	Status      ShowtimeStatus `gorm:"type:showtime_status;not null;default:'scheduled'"`
	IsPublished bool           `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time

	Movie  Movie
	Room   Room
	Prices []ShowtimePrice
}

type ShowtimeState string

const (
	ShowtimeStateScheduled ShowtimeState = "scheduled"
	ShowtimeStatePlaying   ShowtimeState = "playing"
	ShowtimeStateFinished  ShowtimeState = "finished"
	ShowtimeStateCancelled ShowtimeState = "cancelled"
)

func (s *Showtime) State(now time.Time) ShowtimeState {
	switch {
	case s.Status == ShowtimeStatusCancelled:
		return ShowtimeStateCancelled
	case !s.EndsAt.After(now):
		return ShowtimeStateFinished
	case !s.StartsAt.After(now):
		return ShowtimeStatePlaying
	default:
		return ShowtimeStateScheduled
	}
}

func (s *Showtime) Editable(now time.Time) bool {
	return s.State(now) == ShowtimeStateScheduled
}

const counterSaleGrace = 30 * time.Minute

func (s *Showtime) CounterOpen(now time.Time) bool {
	return s.IsPublished && s.Status == ShowtimeStatusScheduled && now.Before(s.StartsAt.Add(counterSaleGrace))
}
