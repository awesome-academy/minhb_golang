package dto

import (
	"sort"
	"time"

	"cinema-booking/internal/models"
)

type ScheduleQuery struct {
	Date string `json:"date" query:"date" validate:"omitempty,datetime=2006-01-02"`
}

type RoomResponse struct {
	ID   int64  `json:"id" example:"1"`
	Name string `json:"name" example:"Room 1"`
}

type ShowtimeSummary struct {
	ID        int64        `json:"id" example:"1"`
	StartsAt  time.Time    `json:"startsAt" example:"2026-09-19T12:00:00Z"`
	EndsAt    time.Time    `json:"endsAt" example:"2026-09-19T14:30:00Z"`
	Format    string       `json:"format" example:"2D"`
	Room      RoomResponse `json:"room"`
	PriceFrom string       `json:"priceFrom" example:"75000.00"`
}

type TheaterSchedule struct {
	TheaterResponse
	Showtimes []ShowtimeSummary `json:"showtimes"`
}

type MovieScheduleResponse struct {
	Movie    MovieSummary      `json:"movie"`
	Date     string            `json:"date" example:"2026-09-19"`
	Theaters []TheaterSchedule `json:"theaters"`
}

type MovieSchedule struct {
	MovieSummary
	Showtimes []ShowtimeSummary `json:"showtimes"`
}

type TheaterScheduleResponse struct {
	Theater TheaterResponse `json:"theater"`
	Date    string          `json:"date" example:"2026-09-19"`
	Movies  []MovieSchedule `json:"movies"`
}

func NewShowtimeSummary(showtime *models.Showtime) ShowtimeSummary {
	return ShowtimeSummary{
		ID:        showtime.ID,
		StartsAt:  showtime.StartsAt.UTC(),
		EndsAt:    showtime.EndsAt.UTC(),
		Format:    string(showtime.Format),
		Room:      RoomResponse{ID: showtime.Room.ID, Name: showtime.Room.Name},
		PriceFrom: minPrice(showtime.Prices),
	}
}

func NewMovieScheduleResponse(movie *models.Movie, date string, showtimes []models.Showtime) MovieScheduleResponse {
	groups := make([]TheaterSchedule, 0)
	index := make(map[int64]int)
	for i := range showtimes {
		showtime := &showtimes[i]
		pos, ok := index[showtime.Room.TheaterID]
		if !ok {
			pos = len(groups)
			index[showtime.Room.TheaterID] = pos
			groups = append(groups, TheaterSchedule{
				TheaterResponse: NewTheaterResponse(&showtime.Room.Theater),
				Showtimes:       make([]ShowtimeSummary, 0),
			})
		}
		groups[pos].Showtimes = append(groups[pos].Showtimes, NewShowtimeSummary(showtime))
	}
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].Name != groups[j].Name {
			return groups[i].Name < groups[j].Name
		}
		return groups[i].ID < groups[j].ID
	})
	return MovieScheduleResponse{Movie: NewMovieSummary(movie), Date: date, Theaters: groups}
}

func NewTheaterScheduleResponse(theater *models.Theater, date string, showtimes []models.Showtime) TheaterScheduleResponse {
	groups := make([]MovieSchedule, 0)
	index := make(map[int64]int)
	for i := range showtimes {
		showtime := &showtimes[i]
		pos, ok := index[showtime.MovieID]
		if !ok {
			pos = len(groups)
			index[showtime.MovieID] = pos
			groups = append(groups, MovieSchedule{
				MovieSummary: NewMovieSummary(&showtime.Movie),
				Showtimes:    make([]ShowtimeSummary, 0),
			})
		}
		groups[pos].Showtimes = append(groups[pos].Showtimes, NewShowtimeSummary(showtime))
	}
	sort.SliceStable(groups, func(i, j int) bool {
		if groups[i].Title != groups[j].Title {
			return groups[i].Title < groups[j].Title
		}
		return groups[i].ID < groups[j].ID
	})
	return TheaterScheduleResponse{Theater: NewTheaterResponse(theater), Date: date, Movies: groups}
}

func minPrice(prices []models.ShowtimePrice) string {
	if len(prices) == 0 {
		return ""
	}
	lowest := prices[0].Price
	for _, price := range prices[1:] {
		if price.Price.LessThan(lowest) {
			lowest = price.Price
		}
	}
	return lowest.StringFixed(2)
}
