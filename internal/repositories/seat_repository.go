package repositories

import (
	"context"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
)

const seatInsertBatchSize = 500

type SeatCount struct {
	RoomID int64
	Total  int64
	Active int64
}

type RowSeatType struct {
	RowLabel   string
	SeatTypeID int64
}

type SeatRepository interface {
	ListByRoom(ctx context.Context, roomID int64) ([]models.Seat, error)
	SeatCounts(ctx context.Context, roomIDs []int64) (map[int64]SeatCount, error)
	HasShowtimes(ctx context.Context, roomID int64) (bool, error)
	Regenerate(ctx context.Context, roomID int64, rows, seatsPerRow int, seatTypeID int64) (int, error)
	ChangeRowTypes(ctx context.Context, roomID int64, rowTypes []RowSeatType) (int, error)
	ChangeStatus(ctx context.Context, roomID, seatID int64) (*models.Seat, error)
}

type seatRepository struct {
	db *gorm.DB
}

func NewSeatRepository(db *gorm.DB) SeatRepository {
	return &seatRepository{db: db}
}

func (r *seatRepository) ListByRoom(ctx context.Context, roomID int64) ([]models.Seat, error) {
	var seats []models.Seat
	err := r.db.WithContext(ctx).Preload("SeatType").
		Where("room_id = ?", roomID).
		Order("length(row_label), row_label, seat_number").
		Find(&seats).Error
	if err != nil {
		return nil, err
	}
	return seats, nil
}

func (r *seatRepository) SeatCounts(ctx context.Context, roomIDs []int64) (map[int64]SeatCount, error) {
	counts := make(map[int64]SeatCount, len(roomIDs))
	if len(roomIDs) == 0 {
		return counts, nil
	}
	var rows []SeatCount
	err := r.db.WithContext(ctx).Model(&models.Seat{}).
		Select("room_id, count(*) AS total, count(*) FILTER (WHERE is_active) AS active").
		Where("room_id IN ?", roomIDs).
		Group("room_id").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		counts[row.RoomID] = row
	}
	return counts, nil
}

func (r *seatRepository) HasShowtimes(ctx context.Context, roomID int64) (bool, error) {
	return hasShowtimes(r.db.WithContext(ctx), roomID)
}

func (r *seatRepository) Regenerate(ctx context.Context, roomID int64, rows, seatsPerRow int, seatTypeID int64) (int, error) {
	seats := buildSeats(roomID, rows, seatsPerRow, seatTypeID)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&models.Room{}, "id = ?", roomID).Error; err != nil {
			return err
		}
		locked, err := hasShowtimes(tx, roomID)
		if err != nil {
			return err
		}
		if locked {
			return apperrors.ErrRoomHasShowtimes
		}
		if err := tx.Where("room_id = ?", roomID).Delete(&models.Seat{}).Error; err != nil {
			return err
		}
		return tx.Omit(clause.Associations).CreateInBatches(&seats, seatInsertBatchSize).Error
	})
	if err != nil {
		return 0, err
	}
	return len(seats), nil
}

func (r *seatRepository) ChangeRowTypes(ctx context.Context, roomID int64, rowTypes []RowSeatType) (int, error) {
	if len(rowTypes) == 0 {
		return 0, nil
	}
	values := make([]string, 0, len(rowTypes))
	args := make([]any, 0, 2*len(rowTypes)+1)
	for _, rowType := range rowTypes {
		values = append(values, "(?::text, ?::bigint)")
		args = append(args, rowType.RowLabel, rowType.SeatTypeID)
	}
	args = append(args, roomID)
	query := `WITH updated AS (
		UPDATE seats AS s SET seat_type_id = v.seat_type_id, updated_at = now()
		FROM (VALUES ` + strings.Join(values, ", ") + `) AS v(row_label, seat_type_id)
		WHERE s.room_id = ? AND s.row_label = v.row_label AND s.seat_type_id <> v.seat_type_id
		RETURNING s.row_label
	) SELECT count(DISTINCT row_label) FROM updated`
	var changed int64
	if err := r.db.WithContext(ctx).Raw(query, args...).Scan(&changed).Error; err != nil {
		return 0, err
	}
	return int(changed), nil
}

func (r *seatRepository) ChangeStatus(ctx context.Context, roomID, seatID int64) (*models.Seat, error) {
	var seat models.Seat
	result := r.db.WithContext(ctx).Model(&seat).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "is_active"}}}).
		Where("id = ? AND room_id = ?", seatID, roomID).
		Update("is_active", gorm.Expr("NOT is_active"))
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &seat, nil
}

func hasShowtimes(db *gorm.DB, roomID int64) (bool, error) {
	var count int64
	err := db.Model(&models.Showtime{}).Where("room_id = ?", roomID).Count(&count).Error
	return count > 0, err
}

func buildSeats(roomID int64, rows, seatsPerRow int, seatTypeID int64) []models.Seat {
	seats := make([]models.Seat, 0, rows*seatsPerRow)
	for row := 0; row < rows; row++ {
		label := rowLabel(row)
		for number := 1; number <= seatsPerRow; number++ {
			seats = append(seats, models.Seat{
				RoomID:     roomID,
				RowLabel:   label,
				SeatNumber: int16(number),
				SeatTypeID: seatTypeID,
				IsActive:   true,
			})
		}
	}
	return seats
}

func rowLabel(index int) string {
	label := ""
	for index >= 0 {
		label = string(rune('A'+index%26)) + label
		index = index/26 - 1
	}
	return label
}
