package repositories

import (
	"context"

	"gorm.io/gorm"

	"cinema-booking/internal/models"
)

var activeTicketStatuses = []models.TicketStatus{models.TicketStatusHeld, models.TicketStatusPaid}

type TicketRepository interface {
	ActiveSeatStatuses(ctx context.Context, showtimeID int64) (map[int64]models.TicketStatus, error)
}

type seatTicket struct {
	SeatID int64
	Status models.TicketStatus
}

type ticketRepository struct {
	db *gorm.DB
}

func NewTicketRepository(db *gorm.DB) TicketRepository {
	return &ticketRepository{db: db}
}

func (r *ticketRepository) ActiveSeatStatuses(ctx context.Context, showtimeID int64) (map[int64]models.TicketStatus, error) {
	var rows []seatTicket
	err := r.db.WithContext(ctx).Model(&models.Ticket{}).
		Select("seat_id, status").
		Where("showtime_id = ? AND status IN ?", showtimeID, activeTicketStatuses).
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	statuses := make(map[int64]models.TicketStatus, len(rows))
	for _, row := range rows {
		statuses[row.SeatID] = row.Status
	}
	return statuses, nil
}
