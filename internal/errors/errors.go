package errors

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid admin credentials")
	ErrSessionNotFound    = errors.New("admin session not found")
	ErrMovieGenreInvalid  = errors.New("movie genre does not exist")
	ErrMovieCastInvalid   = errors.New("movie cast member missing name")
	ErrMovieHasShowtimes  = errors.New("movie has upcoming showtimes")
	ErrRecordModified     = errors.New("record was modified by someone else")
	ErrRecordTokenInvalid = errors.New("record version token is missing or invalid")
	ErrRoomNameTaken      = errors.New("room name already exists in this theater")
	ErrRoomHasShowtimes   = errors.New("room has showtimes")
	ErrSeatTypeInvalid    = errors.New("seat type does not exist")
	ErrRowTypesMismatch   = errors.New("row labels and seat types count mismatch")
)
