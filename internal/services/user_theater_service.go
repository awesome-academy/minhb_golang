package services

import (
	"context"

	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
)

type UserTheaterService interface {
	List(ctx context.Context, city string) ([]models.Theater, error)
	Get(ctx context.Context, id int64) (*models.Theater, error)
}

type userTheaterService struct {
	theaters repositories.TheaterRepository
}

func NewUserTheaterService(theaters repositories.TheaterRepository) UserTheaterService {
	return &userTheaterService{theaters: theaters}
}

func (s *userTheaterService) List(ctx context.Context, city string) ([]models.Theater, error) {
	return s.theaters.ListPublic(ctx, city)
}

func (s *userTheaterService) Get(ctx context.Context, id int64) (*models.Theater, error) {
	return s.theaters.FindPublic(ctx, id)
}
