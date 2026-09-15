package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
)

type TheaterRepository interface {
	List(ctx context.Context, search string, offset, limit int) ([]models.Theater, int64, error)
	FindByID(ctx context.Context, id int64) (*models.Theater, error)
	Create(ctx context.Context, theater *models.Theater) error
	Update(ctx context.Context, theater *models.Theater, expectedUpdatedAt time.Time) error
	ChangeStatus(ctx context.Context, id int64) (bool, error)
}

type theaterRepository struct {
	db *gorm.DB
}

func NewTheaterRepository(db *gorm.DB) TheaterRepository {
	return &theaterRepository{db: db}
}

func (r *theaterRepository) List(ctx context.Context, search string, offset, limit int) ([]models.Theater, int64, error) {
	base := r.db.WithContext(ctx).Model(&models.Theater{})
	if search != "" {
		base = base.Where("name ILIKE ?", "%"+search+"%")
	}
	base = base.Session(&gorm.Session{})
	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var theaters []models.Theater
	if err := base.Order("created_at DESC, id DESC").Offset(offset).Limit(limit).Find(&theaters).Error; err != nil {
		return nil, 0, err
	}
	return theaters, total, nil
}

func (r *theaterRepository) FindByID(ctx context.Context, id int64) (*models.Theater, error) {
	var theater models.Theater
	if err := r.db.WithContext(ctx).First(&theater, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &theater, nil
}

func (r *theaterRepository) Create(ctx context.Context, theater *models.Theater) error {
	return r.db.WithContext(ctx).Omit(clause.Associations).Create(theater).Error
}

func (r *theaterRepository) Update(ctx context.Context, theater *models.Theater, expectedUpdatedAt time.Time) error {
	result := r.db.WithContext(ctx).Model(theater).Omit(clause.Associations).Select("*").
		Where("updated_at = ?", expectedUpdatedAt).Updates(theater)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		if err := r.db.WithContext(ctx).First(&models.Theater{}, "id = ?", theater.ID).Error; err != nil {
			return err
		}
		return apperrors.ErrRecordModified
	}
	return nil
}

func (r *theaterRepository) ChangeStatus(ctx context.Context, id int64) (bool, error) {
	var theater models.Theater
	result := r.db.WithContext(ctx).Model(&theater).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "is_active"}}}).
		Where("id = ?", id).
		Update("is_active", gorm.Expr("NOT is_active"))
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, gorm.ErrRecordNotFound
	}
	return theater.IsActive, nil
}
