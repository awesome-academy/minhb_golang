package repositories

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
)

type RoomRepository interface {
	ListByTheater(ctx context.Context, theaterID int64) ([]models.Room, error)
	ListActive(ctx context.Context) ([]models.Room, error)
	FindByID(ctx context.Context, id int64) (*models.Room, error)
	Create(ctx context.Context, room *models.Room) error
	Update(ctx context.Context, room *models.Room, expectedUpdatedAt time.Time) error
	ChangeStatus(ctx context.Context, id int64) (bool, error)
	NameExists(ctx context.Context, theaterID int64, name string, excludeID int64) (bool, error)
}

type roomRepository struct {
	db *gorm.DB
}

func NewRoomRepository(db *gorm.DB) RoomRepository {
	return &roomRepository{db: db}
}

func (r *roomRepository) ListByTheater(ctx context.Context, theaterID int64) ([]models.Room, error) {
	var rooms []models.Room
	if err := r.db.WithContext(ctx).Where("theater_id = ?", theaterID).Order("name, id").Find(&rooms).Error; err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *roomRepository) ListActive(ctx context.Context) ([]models.Room, error) {
	var rooms []models.Room
	err := r.db.WithContext(ctx).Joins("Theater").Where("rooms.is_active").
		Order(`"Theater".name, rooms.name, rooms.id`).Find(&rooms).Error
	if err != nil {
		return nil, err
	}
	return rooms, nil
}

func (r *roomRepository) FindByID(ctx context.Context, id int64) (*models.Room, error) {
	var room models.Room
	if err := r.db.WithContext(ctx).Preload("Theater").First(&room, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *roomRepository) Create(ctx context.Context, room *models.Room) error {
	return roomWriteError(r.db.WithContext(ctx).Omit(clause.Associations).Create(room).Error)
}

func (r *roomRepository) Update(ctx context.Context, room *models.Room, expectedUpdatedAt time.Time) error {
	result := r.db.WithContext(ctx).Model(room).Omit(clause.Associations).Select("*").
		Where("updated_at = ?", expectedUpdatedAt).Updates(room)
	if result.Error != nil {
		return roomWriteError(result.Error)
	}
	if result.RowsAffected == 0 {
		if err := r.db.WithContext(ctx).First(&models.Room{}, "id = ?", room.ID).Error; err != nil {
			return err
		}
		return apperrors.ErrRecordModified
	}
	return nil
}

func (r *roomRepository) ChangeStatus(ctx context.Context, id int64) (bool, error) {
	var room models.Room
	result := r.db.WithContext(ctx).Model(&room).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "is_active"}}}).
		Where("id = ?", id).
		Update("is_active", gorm.Expr("NOT is_active"))
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, gorm.ErrRecordNotFound
	}
	return room.IsActive, nil
}

func (r *roomRepository) NameExists(ctx context.Context, theaterID int64, name string, excludeID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Unscoped().Model(&models.Room{}).
		Where("theater_id = ? AND lower(name) = lower(?) AND id <> ?", theaterID, name, excludeID).
		Count(&count).Error
	return count > 0, err
}

func roomWriteError(err error) error {
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return apperrors.ErrRoomNameTaken
	}
	return err
}
