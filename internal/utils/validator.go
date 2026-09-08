package utils

import (
	"errors"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
)

type RequestValidator struct {
	validate *validator.Validate
}

func NewRequestValidator() *RequestValidator {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "" || name == "-" {
			return field.Name
		}
		return name
	})
	return &RequestValidator{validate: v}
}

func (rv *RequestValidator) Validate(i any) error {
	err := rv.validate.Struct(i)
	if err == nil {
		return nil
	}

	var fieldErrors validator.ValidationErrors
	if !errors.As(err, &fieldErrors) {
		return APIErrorFrom(http.StatusBadRequest, "request validation failed", err)
	}

	details := make([]ValidationFieldError, 0, len(fieldErrors))
	for _, fe := range fieldErrors {
		details = append(details, ValidationFieldError{
			Field:   fe.Field(),
			Message: validationMessage(fe.Field(), fe),
		})
	}

	if len(details) == 0 {
		return APIErrorFrom(http.StatusBadRequest, "request validation failed", err)
	}

	return ValidationError{
		ErrorCode:    http.StatusBadRequest,
		ErrorMessage: details[0].Message,
		Errors:       details,
	}
}

func BindAndValidate(c *echo.Context, request any) error {
	if err := c.Bind(request); err != nil {
		return APIErrorFrom(http.StatusBadRequest, "invalid request body", err)
	}
	return c.Validate(request)
}

func validationMessage(field string, fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "max":
		return field + " must be at most " + fe.Param()
	case "min":
		return field + " must be at least " + fe.Param()
	case "len":
		return field + " must have length " + fe.Param()
	case "oneof":
		return field + " must be one of: " + fe.Param()
	case "gt", "gte", "lt", "lte":
		return field + " must be " + fe.Tag() + " " + fe.Param()
	default:
		return field + " failed validation rule " + fe.Tag()
	}
}
