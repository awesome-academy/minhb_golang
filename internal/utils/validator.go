package utils

import (
	"errors"
	"net/http"
	"reflect"
	"strings"
	"time"

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
	if err := v.RegisterValidation("pastdate", isPastDate); err != nil {
		panic(err)
	}
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
			Message: validationMessage(fieldLabel(i, fe), fe),
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
	trimStringFields(request)
	return c.Validate(request)
}

func trimStringFields(request any) {
	v := reflect.ValueOf(request)
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		return
	}
	v = v.Elem()
	for i := 0; i < v.NumField(); i++ {
		field := v.Field(i)
		if field.Kind() == reflect.String && field.CanSet() {
			field.SetString(strings.TrimSpace(field.String()))
		}
	}
}

func fieldLabel(request any, fe validator.FieldError) string {
	t := reflect.TypeOf(request)
	for t != nil && t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	name, _, _ := strings.Cut(fe.StructField(), "[")
	if t != nil && t.Kind() == reflect.Struct {
		if field, ok := t.FieldByName(name); ok && field.Tag.Get("label") != "" {
			return field.Tag.Get("label")
		}
	}
	return fe.Field()
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
	case "url":
		return field + " must be a valid URL"
	case "datetime":
		if strings.Contains(fe.Param(), "15:04") {
			return field + " must be a date and time"
		}
		return field + " must be a date in format YYYY-MM-DD"
	case "pastdate":
		return field + " must be in the past"
	case "oneof":
		return field + " must be one of: " + fe.Param()
	case "gt", "gte", "lt", "lte":
		return field + " must be " + fe.Tag() + " " + fe.Param()
	default:
		return field + " failed validation rule " + fe.Tag()
	}
}

func isPastDate(fl validator.FieldLevel) bool {
	value, ok := fl.Field().Interface().(string)
	if !ok {
		return false
	}
	date, err := time.ParseInLocation(time.DateOnly, value, time.UTC)
	if err != nil {
		return true
	}
	today := time.Now().UTC().Truncate(24 * time.Hour)
	return date.Before(today)
}
