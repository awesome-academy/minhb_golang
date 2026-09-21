package services

import (
	"context"
	"time"

	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
	"cinema-booking/internal/utils"
)

type Schedule struct {
	Date      string
	Showtimes []models.Showtime
}

type ShowtimeSeatMap struct {
	Showtime *models.Showtime
	Seats    []models.Seat
	Statuses map[int64]models.TicketStatus
}

type UserShowtimeService interface {
	ByMovie(ctx context.Context, movieID int64, date string) (*models.Movie, *Schedule, error)
	ByTheater(ctx context.Context, theaterID int64, date string) (*models.Theater, *Schedule, error)
	SeatMap(ctx context.Context, id int64) (*ShowtimeSeatMap, error)
}

type userShowtimeService struct {
	movies    repositories.MovieRepository
	theaters  repositories.TheaterRepository
	showtimes repositories.ShowtimeRepository
	seats     repositories.SeatRepository
	tickets   repositories.TicketRepository
}

func NewUserShowtimeService(
	movies repositories.MovieRepository,
	theaters repositories.TheaterRepository,
	showtimes repositories.ShowtimeRepository,
	seats repositories.SeatRepository,
	tickets repositories.TicketRepository,
) UserShowtimeService {
	return &userShowtimeService{movies: movies, theaters: theaters, showtimes: showtimes, seats: seats, tickets: tickets}
}

func (s *userShowtimeService) ByMovie(ctx context.Context, movieID int64, date string) (*models.Movie, *Schedule, error) {
	movie, err := s.movies.FindByID(ctx, movieID)
	if err != nil {
		return nil, nil, err
	}
	schedule, err := s.schedule(ctx, repositories.PublicShowtimeFilter{MovieID: movieID}, date)
	if err != nil {
		return nil, nil, err
	}
	return movie, schedule, nil
}

func (s *userShowtimeService) ByTheater(ctx context.Context, theaterID int64, date string) (*models.Theater, *Schedule, error) {
	theater, err := s.theaters.FindPublic(ctx, theaterID)
	if err != nil {
		return nil, nil, err
	}
	schedule, err := s.schedule(ctx, repositories.PublicShowtimeFilter{TheaterID: theaterID}, date)
	if err != nil {
		return nil, nil, err
	}
	return theater, schedule, nil
}

func (s *userShowtimeService) SeatMap(ctx context.Context, id int64) (*ShowtimeSeatMap, error) {
	showtime, err := s.showtimes.FindPublic(ctx, id)
	if err != nil {
		return nil, err
	}
	seats, err := s.seats.ListByRoom(ctx, showtime.RoomID)
	if err != nil {
		return nil, err
	}
	statuses, err := s.tickets.ActiveSeatStatuses(ctx, id)
	if err != nil {
		return nil, err
	}
	return &ShowtimeSeatMap{Showtime: showtime, Seats: seats, Statuses: statuses}, nil
}

func (s *userShowtimeService) schedule(ctx context.Context, filter repositories.PublicShowtimeFilter, date string) (*Schedule, error) {
	day := scheduleDay(date)
	filter.From = day
	filter.To = day.AddDate(0, 0, 1)
	showtimes, err := s.showtimes.ListPublic(ctx, filter)
	if err != nil {
		return nil, err
	}
	return &Schedule{Date: day.Format(time.DateOnly), Showtimes: showtimes}, nil
}

func scheduleDay(date string) time.Time {
	if date == "" {
		now := time.Now().In(utils.Location)
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, utils.Location)
	}
	day, _ := time.ParseInLocation(time.DateOnly, date, utils.Location)
	return day
}
