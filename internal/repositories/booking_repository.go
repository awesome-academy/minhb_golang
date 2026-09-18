package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/shopspring/decimal"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
)

const (
	bookingCutoff                = 30 * time.Minute
	expiredHoldsOnSeatsCondition = "status = ? AND expires_at <= now() AND id IN (SELECT booking_id FROM tickets WHERE showtime_id = ? AND seat_id IN ? AND status = ?)"
)

type CreateBookingInput struct {
	UserID     int64
	ShowtimeID int64
	SeatIDs    []int64
	Code       string
}

type BookingRepository interface {
	Create(ctx context.Context, input CreateBookingInput) (*models.Booking, error)
	ExpireHolds(ctx context.Context) (int64, error)
}

type bookingRepository struct {
	db *gorm.DB
}

func NewBookingRepository(db *gorm.DB) BookingRepository {
	return &bookingRepository{db: db}
}

func (r *bookingRepository) Create(ctx context.Context, input CreateBookingInput) (*models.Booking, error) {
	var booking *models.Booking
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		showtime, err := lockPublicShowtime(tx, input.ShowtimeID)
		if err != nil {
			return err
		}
		if !showtime.StartsAt.After(time.Now().Add(bookingCutoff)) {
			return apperrors.ErrBookingTooLate
		}
		if err := lockUser(tx, input.UserID); err != nil {
			return err
		}
		if err := expireHoldsOnSeats(tx, input.ShowtimeID, input.SeatIDs); err != nil {
			return err
		}
		pending, err := hasPendingBooking(tx, input.UserID, input.ShowtimeID)
		if err != nil {
			return err
		}
		if pending {
			return apperrors.ErrBookingPendingExists
		}
		tickets, subtotal, err := buildTickets(tx, showtime, input.SeatIDs)
		if err != nil {
			return err
		}
		expiresAt := showtime.StartsAt.Add(-bookingCutoff)
		booking = &models.Booking{
			Code:           input.Code,
			UserID:         input.UserID,
			ShowtimeID:     input.ShowtimeID,
			Status:         models.BookingStatusPending,
			Subtotal:       subtotal,
			DiscountAmount: decimal.Zero,
			ExpiresAt:      &expiresAt,
		}
		if err := tx.Omit(clause.Associations).Create(booking).Error; err != nil {
			return err
		}
		for i := range tickets {
			tickets[i].BookingID = booking.ID
		}
		if err := tx.Omit(clause.Associations).Create(&tickets).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return apperrors.ErrSeatsTaken
			}
			return err
		}
		booking.Tickets = tickets
		return nil
	})
	if err != nil {
		return nil, err
	}
	return booking, nil
}

func (r *bookingRepository) ExpireHolds(ctx context.Context) (int64, error) {
	var expired int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var ids []int64
		err := tx.Model(&models.Booking{}).
			Where("status = ? AND expires_at <= now()", models.BookingStatusPending).
			Pluck("id", &ids).Error
		if err != nil || len(ids) == 0 {
			return err
		}
		if err := expireBookings(tx, ids); err != nil {
			return err
		}
		expired = int64(len(ids))
		return nil
	})
	return expired, err
}

func lockPublicShowtime(tx *gorm.DB, id int64) (*models.Showtime, error) {
	var showtime models.Showtime
	err := publicShowtimes(tx).
		Clauses(clause.Locking{Strength: clause.LockingStrengthShare}).
		First(&showtime, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &showtime, nil
}

func lockUser(tx *gorm.DB, id int64) error {
	return tx.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate}).
		Select("id").First(&models.User{}, "id = ?", id).Error
}

func expireHoldsOnSeats(tx *gorm.DB, showtimeID int64, seatIDs []int64) error {
	ids, err := bookingIDs(tx, showtimeID, expiredHoldsOnSeatsCondition,
		models.BookingStatusPending, showtimeID, seatIDs, models.TicketStatusHeld)
	if err != nil {
		return err
	}
	return expireBookings(tx, ids)
}

func expireBookings(tx *gorm.DB, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	err := tx.Model(&models.Booking{}).
		Where("id IN ? AND status = ?", ids, models.BookingStatusPending).
		Update("status", models.BookingStatusExpired).Error
	if err != nil {
		return err
	}
	return releaseTickets(tx, ids)
}

func hasPendingBooking(tx *gorm.DB, userID, showtimeID int64) (bool, error) {
	var count int64
	err := tx.Model(&models.Booking{}).
		Where("user_id = ? AND showtime_id = ? AND status = ? AND expires_at > now()",
			userID, showtimeID, models.BookingStatusPending).
		Count(&count).Error
	return count > 0, err
}

func buildTickets(tx *gorm.DB, showtime *models.Showtime, seatIDs []int64) ([]models.Ticket, decimal.Decimal, error) {
	var seats []models.Seat
	err := tx.Where("id IN ? AND room_id = ? AND is_active", seatIDs, showtime.RoomID).
		Order("id").Find(&seats).Error
	if err != nil {
		return nil, decimal.Zero, err
	}
	if len(seats) != len(seatIDs) {
		return nil, decimal.Zero, apperrors.ErrSeatsInvalid
	}
	var prices []models.ShowtimePrice
	if err := tx.Where("showtime_id = ?", showtime.ID).Find(&prices).Error; err != nil {
		return nil, decimal.Zero, err
	}
	priceByType := make(map[int64]decimal.Decimal, len(prices))
	for _, price := range prices {
		priceByType[price.SeatTypeID] = price.Price
	}
	tickets := make([]models.Ticket, 0, len(seats))
	subtotal := decimal.Zero
	for _, seat := range seats {
		price, ok := priceByType[seat.SeatTypeID]
		if !ok {
			return nil, decimal.Zero, apperrors.ErrSeatsInvalid
		}
		tickets = append(tickets, models.Ticket{
			ShowtimeID: showtime.ID,
			SeatID:     seat.ID,
			Price:      price,
			Status:     models.TicketStatusHeld,
		})
		subtotal = subtotal.Add(price)
	}
	return tickets, subtotal, nil
}
