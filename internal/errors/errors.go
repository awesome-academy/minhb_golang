package errors

import (
	"errors"
	"time"
)

var (
	ErrInvalidCredentials   = errors.New("invalid admin credentials")
	ErrSessionNotFound      = errors.New("admin session not found")
	ErrMovieGenreInvalid    = errors.New("movie genre does not exist")
	ErrMovieCastInvalid     = errors.New("movie cast member missing name")
	ErrMovieHasShowtimes    = errors.New("movie has upcoming showtimes")
	ErrRecordModified       = errors.New("record was modified by someone else")
	ErrRecordTokenInvalid   = errors.New("record version token is missing or invalid")
	ErrRoomNameTaken        = errors.New("room name already exists in this theater")
	ErrRoomHasShowtimes     = errors.New("room has showtimes")
	ErrSeatTypeInvalid      = errors.New("seat type does not exist")
	ErrRowTypesMismatch     = errors.New("row labels and seat types count mismatch")
	ErrShowtimeMovieInvalid = errors.New("movie is not available for showtimes")
	ErrShowtimeRoomInvalid  = errors.New("room is not available for showtimes")
	ErrRoomHasNoSeats       = errors.New("room has no seats")
	ErrShowtimeInPast       = errors.New("showtime must start in the future")
	ErrShowtimeEndsTooEarly = errors.New("showtime ends before the movie ends")
	ErrShowtimeNotEditable  = errors.New("showtime can no longer be edited")
	ErrShowtimeLocked       = errors.New("showtime has bookings, movie/room/time are locked")
	ErrShowtimeHasBookings  = errors.New("showtime has bookings")
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
