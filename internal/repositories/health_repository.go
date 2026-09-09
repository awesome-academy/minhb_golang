package repositories

import (
	"context"

	"gorm.io/gorm"
)

type HealthRepository interface {
	Ping(ctx context.Context) error
}

type healthRepository struct {
	db *gorm.DB
}

func NewHealthRepository(db *gorm.DB) HealthRepository {
	return &healthRepository{db: db}
}

func (r *healthRepository) Ping(ctx context.Context) error {
	return r.db.WithContext(ctx).Exec("SELECT 1").Error
}
