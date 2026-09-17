package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
)

type MovieRepository interface {
	List(ctx context.Context, search string, offset, limit int) ([]models.Movie, int64, error)
	ListAvailable(ctx context.Context) ([]models.Movie, error)
	FindByID(ctx context.Context, id int64) (*models.Movie, error)
	Create(ctx context.Context, movie *models.Movie, genreIDs []int64) error
	Update(ctx context.Context, movie *models.Movie, genreIDs []int64, expectedUpdatedAt time.Time) error
	Delete(ctx context.Context, id int64) error
}

type movieRepository struct {
	db *gorm.DB
}

func NewMovieRepository(db *gorm.DB) MovieRepository {
	return &movieRepository{db: db}
}

func (r *movieRepository) List(ctx context.Context, search string, offset, limit int) ([]models.Movie, int64, error) {
	base := r.db.WithContext(ctx).Model(&models.Movie{})
	if search != "" {
		base = base.Where("title ILIKE ?", likePattern(search))
	}
	base = base.Session(&gorm.Session{})

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var movies []models.Movie
	err := base.Preload("Genres").Order("created_at DESC, id DESC").Offset(offset).Limit(limit).Find(&movies).Error
	if err != nil {
		return nil, 0, err
	}
	return movies, total, nil
}

func (r *movieRepository) ListAvailable(ctx context.Context) ([]models.Movie, error) {
	var movies []models.Movie
	err := r.db.WithContext(ctx).Where("status <> ?", models.MovieStatusEnded).Order("title, id").Find(&movies).Error
	if err != nil {
		return nil, err
	}
	return movies, nil
}

func (r *movieRepository) FindByID(ctx context.Context, id int64) (*models.Movie, error) {
	var movie models.Movie
	if err := r.db.WithContext(ctx).Preload("Genres").First(&movie, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &movie, nil
}

func (r *movieRepository) Create(ctx context.Context, movie *models.Movie, genreIDs []int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Genres").Create(movie).Error; err != nil {
			return err
		}
		return replaceGenres(tx, movie.ID, genreIDs)
	})
}

func (r *movieRepository) Update(ctx context.Context, movie *models.Movie, genreIDs []int64, expectedUpdatedAt time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(movie).Omit("Genres").Select("*").Where("updated_at = ?", expectedUpdatedAt).Updates(movie)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			if err := tx.First(&models.Movie{}, "id = ?", movie.ID).Error; err != nil {
				return err
			}
			return apperrors.ErrRecordModified
		}
		return replaceGenres(tx, movie.ID, genreIDs)
	})
}

func (r *movieRepository) Delete(ctx context.Context, id int64) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var movie models.Movie
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&movie, "id = ?", id).Error; err != nil {
			return err
		}
		var upcoming int64
		err := tx.Model(&models.Showtime{}).
			Where("movie_id = ? AND status = ? AND starts_at > now()", id, models.ShowtimeStatusScheduled).
			Count(&upcoming).Error
		if err != nil {
			return err
		}
		if upcoming > 0 {
			return apperrors.ErrMovieHasShowtimes
		}
		if err := replaceGenres(tx, id, nil); err != nil {
			return err
		}
		return tx.Delete(&movie).Error
	})
}

func replaceGenres(tx *gorm.DB, movieID int64, genreIDs []int64) error {
	if err := tx.Where("movie_id = ?", movieID).Delete(&models.MovieGenre{}).Error; err != nil {
		return err
	}
	if len(genreIDs) == 0 {
		return nil
	}
	rows := make([]models.MovieGenre, 0, len(genreIDs))
	for _, genreID := range genreIDs {
		rows = append(rows, models.MovieGenre{MovieID: movieID, GenreID: genreID})
	}
	return tx.Create(&rows).Error
}
