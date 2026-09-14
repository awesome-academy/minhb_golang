package errors

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid admin credentials")
	ErrSessionNotFound    = errors.New("admin session not found")
	ErrMovieGenreInvalid  = errors.New("movie genre does not exist")
	ErrMovieCastInvalid   = errors.New("movie cast member missing name")
	ErrMovieHasShowtimes  = errors.New("movie has upcoming showtimes")
	ErrMovieModified      = errors.New("movie was modified by someone else")
)
