package services

import (
	"context"

	"cinema-booking/internal/repositories"
)

type HealthService interface {
	Check(ctx context.Context) error
}

type healthService struct {
	healthRepository repositories.HealthRepository
}

func NewHealthService(healthRepository repositories.HealthRepository) HealthService {
	return &healthService{healthRepository: healthRepository}
}

func (s *healthService) Check(ctx context.Context) error {
	return s.healthRepository.Ping(ctx)
}
