package models

import "time"

type Seat struct {
	ID         int64  `gorm:"primaryKey"`
	RoomID     int64  `gorm:"not null;uniqueIndex:seats_room_id_row_label_seat_number_key"`
	RowLabel   string `gorm:"not null;uniqueIndex:seats_room_id_row_label_seat_number_key"`
	SeatNumber int16  `gorm:"type:smallint;not null;uniqueIndex:seats_room_id_row_label_seat_number_key"`
	SeatTypeID int64  `gorm:"not null"`
	IsActive   bool   `gorm:"not null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time

	SeatType SeatType
}

func CoupleSeatPartner(number int16) int16 {
	if number%2 == 1 {
		return number + 1
	}
	return number - 1
}
