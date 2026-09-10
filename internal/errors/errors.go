package errors

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid admin credentials")
	ErrSessionNotFound    = errors.New("admin session not found")
)
