package errors

import (
	"errors"
	"time"
)

var (
	ErrInvalidCredentials      = errors.New("invalid email or password")
	ErrSessionNotFound         = errors.New("admin session not found")
	ErrEmailTaken              = errors.New("email already exists")
	ErrAccessTokenRevoked      = errors.New("access token has been revoked")
	ErrMovieGenreInvalid       = errors.New("movie genre does not exist")
	ErrMovieCastInvalid        = errors.New("movie cast member missing name")
	ErrMovieHasShowtimes       = errors.New("movie has upcoming showtimes")
	ErrRecordModified          = errors.New("record was modified by someone else")
	ErrRecordTokenInvalid      = errors.New("record version token is missing or invalid")
	ErrRoomNameTaken           = errors.New("room name already exists in this theater")
	ErrRoomHasShowtimes        = errors.New("room has showtimes")
	ErrSeatTypeInvalid         = errors.New("seat type does not exist")
	ErrRowTypesMismatch        = errors.New("row labels and seat types count mismatch")
	ErrShowtimeMovieInvalid    = errors.New("movie is not available for showtimes")
	ErrShowtimeRoomInvalid     = errors.New("room is not available for showtimes")
	ErrRoomHasNoSeats          = errors.New("room has no seats")
	ErrShowtimeInPast          = errors.New("showtime must start in the future")
	ErrShowtimeEndsTooEarly    = errors.New("showtime ends before the movie ends")
	ErrShowtimeNotEditable     = errors.New("showtime can no longer be edited")
	ErrShowtimeLocked          = errors.New("showtime has bookings, movie/room/time are locked")
	ErrShowtimeHasBookings     = errors.New("showtime has bookings")
	ErrBookingTooLate          = errors.New("bookings close 30 minutes before the showtime starts")
	ErrSeatsInvalid            = errors.New("one or more seats are not available for this showtime")
	ErrSeatsTaken              = errors.New("one or more seats have just been taken")
	ErrCoupleSeatsUnpaired     = errors.New("couple seats must be booked in pairs (1-2, 3-4, ...), select both seats of the pair")
	ErrBookingPendingExists    = errors.New("you already have a pending booking for this showtime")
	ErrCounterClosed           = errors.New("counter sales are closed for this showtime")
	ErrBookingCodeMismatch     = errors.New("booking code does not match")
	ErrBookingAlreadyConfirmed = errors.New("booking is already paid")
	ErrBookingCancelled        = errors.New("booking was cancelled with its showtime")
	ErrBookingExpired          = errors.New("booking hold has expired")
)

type FieldErrors map[string]string

func (e FieldErrors) Error() string { return "invalid form fields" }

type ShowtimeOverlapError struct {
	MovieTitle string
	StartsAt   time.Time
	EndsAt     time.Time
}

func (e *ShowtimeOverlapError) Error() string {
	return "showtime overlaps another showtime in this room"
}
