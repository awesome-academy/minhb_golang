package utils

import (
	"encoding/json"
	"errors"
	"net/http"

	"gorm.io/gorm"
)

type ValidationFieldError struct {
	Field   string `json:"field" example:"email"`
	Message string `json:"message" example:"email must be a valid email"`
}

type ValidationError struct {
	ErrorCode    int                    `json:"errorCode" example:"400"`
	ErrorMessage string                 `json:"errorMessage" example:"request validation failed"`
	Errors       []ValidationFieldError `json:"errors"`
}

func (e ValidationError) Error() string { return e.ErrorMessage }

func (e ValidationError) StatusCode() int { return http.StatusBadRequest }

func (e ValidationError) MarshalJSON() ([]byte, error) {
	type validationError ValidationError
	return json.Marshal(validationError(e))
}

type APIErrorResponse struct {
	ErrorCode    int    `json:"errorCode" example:"404"`
	ErrorMessage string `json:"errorMessage" example:"resource not found"`

	cause error
}

func (e APIErrorResponse) Error() string {
	if e.cause != nil {
		return e.ErrorMessage + ": " + e.cause.Error()
	}
	return e.ErrorMessage
}

func (e APIErrorResponse) StatusCode() int { return e.ErrorCode }

func (e APIErrorResponse) Unwrap() error { return e.cause }

func (e APIErrorResponse) MarshalJSON() ([]byte, error) {
	type jsonAPIErrorResponse struct {
		ErrorCode    int    `json:"errorCode"`
		ErrorMessage string `json:"errorMessage"`
	}
	return json.Marshal(jsonAPIErrorResponse{ErrorCode: e.ErrorCode, ErrorMessage: e.ErrorMessage})
}

func APIError(statusCode int, message string) APIErrorResponse {
	return APIErrorResponse{ErrorCode: statusCode, ErrorMessage: message}
}

func APIErrorFrom(statusCode int, message string, cause error) APIErrorResponse {
	return APIErrorResponse{ErrorCode: statusCode, ErrorMessage: message, cause: cause}
}

type statusCoder interface{ StatusCode() int }

func ServiceError(err error) error {
	var coder statusCoder
	if errors.As(err, &coder) {
		return err
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return APIErrorFrom(http.StatusNotFound, "resource not found", err)
	}

	return APIErrorFrom(http.StatusInternalServerError, "internal server error", err)
}
