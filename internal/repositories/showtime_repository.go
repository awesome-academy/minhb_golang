package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
)

type ShowtimeFilter struct {
	TheaterID int64
	From      time.Time
	To        time.Time
	State     models.ShowtimeState
	Published *bool
}

type PublicShowtimeFilter struct {
	MovieID   int64
	TheaterID int64
	From      time.Time
	To        time.Time
}

type ShowtimeRepository interface {
	List(ctx context.Context, filter ShowtimeFilter, offset, limit int) ([]models.Showtime, int64, error)
	ListPublic(ctx context.Context, filter PublicShowtimeFilter) ([]models.Showtime, error)
	FindByID(ctx context.Context, id int64) (*models.Showtime, error)
	FindPublic(ctx context.Context, id int64) (*models.Showtime, error)
	LatestInTheater(ctx context.Context, theaterID int64) (*models.Showtime, error)
	HasActiveBookings(ctx context.Context, id int64) (bool, error)
	ActiveBookingCounts(ctx context.Context, ids []int64) (map[int64]int64, error)
	Create(ctx context.Context, showtime *models.Showtime, prices []models.ShowtimePrice) error
	Update(ctx context.Context, showtime *models.Showtime, prices []models.ShowtimePrice, expectedUpdatedAt time.Time) error
	ChangePublished(ctx context.Context, id int64) (bool, error)
	Cancel(ctx context.Context, id int64) (int64, error)
}

type bookingCount struct {
	ShowtimeID int64
	Count      int64
}

const activeBookingCondition = "(status = ? OR (status = ? AND expires_at > now()))"

type showtimeRepository struct {
	db *gorm.DB
}

func NewShowtimeRepository(db *gorm.DB) ShowtimeRepository {
	return &showtimeRepository{db: db}
}

func (r *showtimeRepository) List(ctx context.Context, filter ShowtimeFilter, offset, limit int) ([]models.Showtime, int64, error) {
	base := r.db.WithContext(ctx).Model(&models.Showtime{})
	if filter.TheaterID > 0 {
		base = base.Where("room_id IN (SELECT id FROM rooms WHERE theater_id = ?)", filter.TheaterID)
	}
	if !filter.From.IsZero() {
		base = base.Where("starts_at >= ?", filter.From)
	}
	if !filter.To.IsZero() {
		base = base.Where("starts_at < ?", filter.To)
	}
	if filter.Published != nil {
		base = base.Where("is_published = ?", *filter.Published)
	}
	switch filter.State {
	case models.ShowtimeStateScheduled:
		base = base.Where("status = ? AND starts_at > now()", models.ShowtimeStatusScheduled)
	case models.ShowtimeStatePlaying:
		base = base.Where("status = ? AND starts_at <= now() AND ends_at > now()", models.ShowtimeStatusScheduled)
	case models.ShowtimeStateFinished:
		base = base.Where("status = ? AND ends_at <= now()", models.ShowtimeStatusScheduled)
	case models.ShowtimeStateCancelled:
		base = base.Where("status = ?", models.ShowtimeStatusCancelled)
	}
	base = base.Session(&gorm.Session{})

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var showtimes []models.Showtime
	err := base.Preload("Movie", unscoped).Preload("Room.Theater").Preload("Prices").
		Order("starts_at DESC, id DESC").Offset(offset).Limit(limit).
		Find(&showtimes).Error
	if err != nil {
		return nil, 0, err
	}
	return showtimes, total, nil
}

func (r *showtimeRepository) ListPublic(ctx context.Context, filter PublicShowtimeFilter) ([]models.Showtime, error) {
	query := publicShowtimes(r.db.WithContext(ctx)).Where("starts_at >= ? AND starts_at < ?", filter.From, filter.To)
	if filter.MovieID > 0 {
		query = query.Where("movie_id = ?", filter.MovieID)
	}
	if filter.TheaterID > 0 {
		query = query.Where("room_id IN (SELECT id FROM rooms WHERE theater_id = ?)", filter.TheaterID)
	}
	var showtimes []models.Showtime
	err := query.Preload("Movie.Genres").Preload("Room.Theater").Preload("Prices").
		Order("starts_at, id").
		Find(&showtimes).Error
	if err != nil {
		return nil, err
	}
	return showtimes, nil
}

func (r *showtimeRepository) FindByID(ctx context.Context, id int64) (*models.Showtime, error) {
	var showtime models.Showtime
	err := r.db.WithContext(ctx).
		Preload("Movie", unscoped).Preload("Room.Theater").Preload("Prices").
		First(&showtime, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &showtime, nil
}

func (r *showtimeRepository) FindPublic(ctx context.Context, id int64) (*models.Showtime, error) {
	var showtime models.Showtime
	err := publicShowtimes(r.db.WithContext(ctx)).
		Preload("Movie.Genres").Preload("Room.Theater").Preload("Prices.SeatType").
		First(&showtime, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &showtime, nil
}

func (r *showtimeRepository) LatestInTheater(ctx context.Context, theaterID int64) (*models.Showtime, error) {
	var showtimes []models.Showtime
	err := r.db.WithContext(ctx).Preload("Prices").
		Where("status = ? AND room_id IN (SELECT id FROM rooms WHERE theater_id = ?)", models.ShowtimeStatusScheduled, theaterID).
		Order("starts_at DESC, id DESC").Limit(1).
		Find(&showtimes).Error
	if err != nil || len(showtimes) == 0 {
		return nil, err
	}
	return &showtimes[0], nil
}

func (r *showtimeRepository) HasActiveBookings(ctx context.Context, id int64) (bool, error) {
	return hasActiveBookings(r.db.WithContext(ctx), id)
}

func (r *showtimeRepository) ActiveBookingCounts(ctx context.Context, ids []int64) (map[int64]int64, error) {
	counts := make(map[int64]int64, len(ids))
	if len(ids) == 0 {
		return counts, nil
	}
	var rows []bookingCount
	err := r.db.WithContext(ctx).Model(&models.Booking{}).
		Select("showtime_id, count(*) AS count").
		Where("showtime_id IN ?", ids).
		Where(activeBookingCondition, models.BookingStatusConfirmed, models.BookingStatusPending).
		Group("showtime_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.ShowtimeID] = row.Count
	}
	return counts, nil
}

func (r *showtimeRepository) Create(ctx context.Context, showtime *models.Showtime, prices []models.ShowtimePrice) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockRoom(tx, showtime.RoomID); err != nil {
			return err
		}
		if err := checkOverlap(tx, showtime, 0); err != nil {
			return err
		}
		if err := tx.Omit(clause.Associations).Create(showtime).Error; err != nil {
			return err
		}
		return replacePrices(tx, showtime.ID, prices)
	})
}

func (r *showtimeRepository) Update(ctx context.Context, showtime *models.Showtime, prices []models.ShowtimePrice, expectedUpdatedAt time.Time) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := lockRoom(tx, showtime.RoomID); err != nil {
			return err
		}
		var current models.Showtime
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, "id = ?", showtime.ID).Error; err != nil {
			return err
		}
		if !current.Editable(time.Now()) {
			return apperrors.ErrShowtimeNotEditable
		}
		changed := scheduleChanged(&current, showtime)
		unpublishing := current.IsPublished && !showtime.IsPublished
		if changed || unpublishing {
			locked, err := hasActiveBookings(tx, showtime.ID)
			if err != nil {
				return err
			}
			if locked && changed {
				return apperrors.ErrShowtimeLocked
			}
			if locked {
				return apperrors.ErrShowtimeHasBookings
			}
		}
		if err := checkOverlap(tx, showtime, showtime.ID); err != nil {
			return err
		}
		result := tx.Model(showtime).Omit(clause.Associations).Select("*").
			Where("updated_at = ?", expectedUpdatedAt).Updates(showtime)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return apperrors.ErrRecordModified
		}
		return replacePrices(tx, showtime.ID, prices)
	})
}

func (r *showtimeRepository) ChangePublished(ctx context.Context, id int64) (bool, error) {
	var published bool
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current models.Showtime
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, "id = ?", id).Error; err != nil {
			return err
		}
		if !current.Editable(time.Now()) {
			return apperrors.ErrShowtimeNotEditable
		}
		if current.IsPublished {
			locked, err := hasActiveBookings(tx, id)
			if err != nil {
				return err
			}
			if locked {
				return apperrors.ErrShowtimeHasBookings
			}
		}
		published = !current.IsPublished
		return tx.Model(&current).Update("is_published", published).Error
	})
	return published, err
}

func (r *showtimeRepository) Cancel(ctx context.Context, id int64) (int64, error) {
	var cancelled int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current models.Showtime
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&current, "id = ?", id).Error; err != nil {
			return err
		}
		if !current.Editable(time.Now()) {
			return apperrors.ErrShowtimeNotEditable
		}
		stale, err := bookingIDs(tx, id, "status = ? AND expires_at <= now()", models.BookingStatusPending)
		if err != nil {
			return err
		}
		active, err := bookingIDs(tx, id, activeBookingCondition, models.BookingStatusConfirmed, models.BookingStatusPending)
		if err != nil {
			return err
		}
		if err := updateBookings(tx, stale, map[string]any{"status": models.BookingStatusExpired}); err != nil {
			return err
		}
		if err := updateBookings(tx, active, map[string]any{
			"status": models.BookingStatusCancelled, "note": "Showtime cancelled by cinema", "expires_at": nil,
		}); err != nil {
			return err
		}
		if err := releaseTickets(tx, append(stale, active...)); err != nil {
			return err
		}
		cancelled = int64(len(active))
		return tx.Model(&current).Update("status", models.ShowtimeStatusCancelled).Error
	})
	return cancelled, err
}

func unscoped(db *gorm.DB) *gorm.DB {
	return db.Unscoped()
}

func publicShowtimes(db *gorm.DB) *gorm.DB {
	return db.Where("status = ? AND is_published AND starts_at > now()", models.ShowtimeStatusScheduled).
		Where("movie_id IN (SELECT id FROM movies WHERE deleted_at IS NULL)").
		Where(`room_id IN (SELECT r.id FROM rooms r JOIN theaters t ON t.id = r.theater_id
			WHERE r.is_active AND r.deleted_at IS NULL AND t.is_active AND t.deleted_at IS NULL)`)
}

func lockRoom(tx *gorm.DB, roomID int64) error {
	return tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&models.Room{}, "id = ?", roomID).Error
}

func hasActiveBookings(db *gorm.DB, showtimeID int64) (bool, error) {
	var count int64
	err := db.Model(&models.Booking{}).Where("showtime_id = ?", showtimeID).
		Where(activeBookingCondition, models.BookingStatusConfirmed, models.BookingStatusPending).
		Count(&count).Error
	return count > 0, err
}

func bookingIDs(tx *gorm.DB, showtimeID int64, condition string, args ...any) ([]int64, error) {
	var ids []int64
	err := tx.Model(&models.Booking{}).Where("showtime_id = ?", showtimeID).Where(condition, args...).Pluck("id", &ids).Error
	return ids, err
}

func updateBookings(tx *gorm.DB, ids []int64, values map[string]any) error {
	if len(ids) == 0 {
		return nil
	}
	return tx.Model(&models.Booking{}).Where("id IN ?", ids).Updates(values).Error
}

func releaseTickets(tx *gorm.DB, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return tx.Model(&models.Ticket{}).
		Where("booking_id IN ? AND status IN ?", ids, activeTicketStatuses).
		Update("status", models.TicketStatusReleased).Error
}

func checkOverlap(tx *gorm.DB, showtime *models.Showtime, excludeID int64) error {
	var overlapping []models.Showtime
	err := tx.Preload("Movie", unscoped).
		Where("room_id = ? AND status = ? AND id <> ? AND starts_at < ? AND ends_at > ?",
			showtime.RoomID, models.ShowtimeStatusScheduled, excludeID, showtime.EndsAt, showtime.StartsAt).
		Order("starts_at").Limit(1).
		Find(&overlapping).Error
	if err != nil || len(overlapping) == 0 {
		return err
	}
	other := overlapping[0]
	return &apperrors.ShowtimeOverlapError{MovieTitle: other.Movie.Title, StartsAt: other.StartsAt, EndsAt: other.EndsAt}
}

func scheduleChanged(current, next *models.Showtime) bool {
	return current.MovieID != next.MovieID || current.RoomID != next.RoomID ||
		!current.StartsAt.Equal(next.StartsAt) || !current.EndsAt.Equal(next.EndsAt)
}

func replacePrices(tx *gorm.DB, showtimeID int64, prices []models.ShowtimePrice) error {
	if err := tx.Where("showtime_id = ?", showtimeID).Delete(&models.ShowtimePrice{}).Error; err != nil {
		return err
	}
	if len(prices) == 0 {
		return nil
	}
	for i := range prices {
		prices[i].ShowtimeID = showtimeID
	}
	return tx.Omit(clause.Associations).Create(&prices).Error
}
