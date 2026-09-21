package services

import (
	"context"
	"strings"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
)

const UserMoviePageSize = 20

type MoviePage struct {
	Movies   []models.Movie
	Total    int64
	Page     int
	PageSize int
}

type UserMovieService interface {
	List(ctx context.Context, query dto.MovieListQuery) (*MoviePage, error)
	Get(ctx context.Context, id int64) (*models.Movie, error)
}

type userMovieService struct {
	movies repositories.MovieRepository
}

func NewUserMovieService(movies repositories.MovieRepository) UserMovieService {
	return &userMovieService{movies: movies}
}

func (s *userMovieService) List(ctx context.Context, query dto.MovieListQuery) (*MoviePage, error) {
	page, pageSize := normalizePage(query.Page, query.PageSize, UserMoviePageSize)
	filter := repositories.MovieFilter{Status: models.MovieStatus(query.Status), Search: strings.TrimSpace(query.Q)}
	movies, total, err := s.movies.ListPublic(ctx, filter, (page-1)*pageSize, pageSize)
	if err != nil {
		return nil, err
	}
	return &MoviePage{Movies: movies, Total: total, Page: page, PageSize: pageSize}, nil
}

func (s *userMovieService) Get(ctx context.Context, id int64) (*models.Movie, error) {
	return s.movies.FindByID(ctx, id)
}

func normalizePage(page, pageSize, defaultPageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultPageSize
	}
	return page, pageSize
}
