package dto

import (
	"time"

	"cinema-booking/internal/models"
)

type MovieListQuery struct {
	Status   string `json:"status" query:"status" validate:"omitempty,oneof=now_showing coming_soon"`
	Q        string `json:"q" query:"q" validate:"omitempty,min=1,max=100"`
	Page     int    `json:"page" query:"page" validate:"omitempty,min=1,max=100000"`
	PageSize int    `json:"pageSize" query:"pageSize" validate:"omitempty,min=1,max=50"`
}

type Pagination struct {
	Page     int   `json:"page" example:"1"`
	PageSize int   `json:"pageSize" example:"20"`
	Total    int64 `json:"total" example:"57"`
}

type GenreResponse struct {
	ID   int64  `json:"id" example:"1"`
	Name string `json:"name" example:"Action"`
	Slug string `json:"slug" example:"action"`
}

type CastMemberResponse struct {
	Name string `json:"name" example:"Leonardo DiCaprio"`
	Role string `json:"role" example:"Cobb"`
}

type MovieSummary struct {
	ID          int64           `json:"id" example:"1"`
	Slug        string          `json:"slug" example:"inception-1a2b3c4d"`
	Title       string          `json:"title" example:"Inception"`
	PosterURL   *string         `json:"posterUrl" example:"https://cdn.example.com/posters/inception.jpg"`
	DurationMin int             `json:"durationMin" example:"148"`
	AgeRating   string          `json:"ageRating" example:"T13"`
	ReleaseDate string          `json:"releaseDate" example:"2026-10-01"`
	Status      string          `json:"status" example:"now_showing"`
	Genres      []GenreResponse `json:"genres"`
}

type MovieDetail struct {
	MovieSummary
	OriginalTitle *string              `json:"originalTitle" example:"Inception"`
	Description   *string              `json:"description" example:"A thief who steals corporate secrets through dream-sharing technology."`
	Language      *string              `json:"language" example:"English"`
	Director      *string              `json:"director" example:"Christopher Nolan"`
	CastMembers   []CastMemberResponse `json:"castMembers"`
	BackdropURL   *string              `json:"backdropUrl" example:"https://cdn.example.com/backdrops/inception.jpg"`
	TrailerURL    *string              `json:"trailerUrl" example:"https://www.youtube.com/watch?v=YoHD9XEInc0"`
}

type MovieListResponse struct {
	Items []MovieSummary `json:"items"`
	Pagination
}

func NewMovieSummary(movie *models.Movie) MovieSummary {
	genres := make([]GenreResponse, 0, len(movie.Genres))
	for _, genre := range movie.Genres {
		genres = append(genres, GenreResponse{ID: genre.ID, Name: genre.Name, Slug: genre.Slug})
	}
	return MovieSummary{
		ID:          movie.ID,
		Slug:        movie.Slug,
		Title:       movie.Title,
		PosterURL:   movie.PosterURL,
		DurationMin: movie.DurationMin,
		AgeRating:   string(movie.AgeRating),
		ReleaseDate: movie.ReleaseDate.Format(time.DateOnly),
		Status:      string(movie.Status),
		Genres:      genres,
	}
}

func NewMovieDetail(movie *models.Movie) MovieDetail {
	cast := make([]CastMemberResponse, 0, len(movie.CastMembers))
	for _, member := range movie.CastMembers {
		cast = append(cast, CastMemberResponse{Name: member.Name, Role: member.Role})
	}
	return MovieDetail{
		MovieSummary:  NewMovieSummary(movie),
		OriginalTitle: movie.OriginalTitle,
		Description:   movie.Description,
		Language:      movie.Language,
		Director:      movie.Director,
		CastMembers:   cast,
		BackdropURL:   movie.BackdropURL,
		TrailerURL:    movie.TrailerURL,
	}
}

func NewMovieListResponse(movies []models.Movie, page, pageSize int, total int64) MovieListResponse {
	items := make([]MovieSummary, 0, len(movies))
	for i := range movies {
		items = append(items, NewMovieSummary(&movies[i]))
	}
	return MovieListResponse{
		Items:      items,
		Pagination: Pagination{Page: page, PageSize: pageSize, Total: total},
	}
}
