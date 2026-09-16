package repositories

import (
	"context"

	"gorm.io/gorm"

	"cinema-booking/internal/models"
)

type GenreRepository interface {
	ListAll(ctx context.Context) ([]models.Genre, error)
	CountByIDs(ctx context.Context, ids []int64) (int64, error)
}

type genreRepository struct {
	db *gorm.DB
}

func NewGenreRepository(db *gorm.DB) GenreRepository {
	return &genreRepository{db: db}
}

func (r *genreRepository) ListAll(ctx context.Context) ([]models.Genre, error) {
	var genres []models.Genre
	if err := r.db.WithContext(ctx).Order("name").Find(&genres).Error; err != nil {
		return nil, err
	}
	return genres, nil
}

func (r *genreRepository) CountByIDs(ctx context.Context, ids []int64) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	var count int64
	err := r.db.WithContext(ctx).Model(&models.Genre{}).Where("id IN ?", ids).Count(&count).Error
	return count, err
}
