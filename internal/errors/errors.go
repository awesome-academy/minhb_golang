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
)
