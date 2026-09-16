package repositories

import (
	"context"

	"gorm.io/gorm"

	"cinema-booking/internal/models"
)

type SeatTypeRepository interface {
	ListAll(ctx context.Context) ([]models.SeatType, error)
}

type seatTypeRepository struct {
	db *gorm.DB
}

func NewSeatTypeRepository(db *gorm.DB) SeatTypeRepository {
	return &seatTypeRepository{db: db}
}

func (r *seatTypeRepository) ListAll(ctx context.Context) ([]models.SeatType, error) {
	var seatTypes []models.SeatType
	if err := r.db.WithContext(ctx).Order("id").Find(&seatTypes).Error; err != nil {
		return nil, err
	}
	return seatTypes, nil
}
