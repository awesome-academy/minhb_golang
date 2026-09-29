package admin

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"

	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/middleware"
	"cinema-booking/internal/utils"
)

const (
	flashCookieName = "admin_flash"
	flashSuccess    = "success"
	flashDanger     = "danger"
)

type Flash struct {
	Kind    string
	Message string
}

type flashCookie struct {
	secure bool
}

func csrfToken(c *echo.Context) string {
	token, _ := c.Get(middleware.CSRFContextKey).(string)
	return token
}

func adminEmail(c *echo.Context) string {
	if user := middleware.AdminUser(c); user != nil {
		return user.Email
	}
	return ""
}

func parseID(c *echo.Context) (int64, error) {
	return parseParamID(c, "id")
}

func parseParamID(c *echo.Context, name string) (int64, error) {
	return utils.ParamID(c, name)
}

func parsePage(raw string) int {
	page, err := strconv.Atoi(raw)
	if err != nil || page < 1 {
		return 1
	}
	return page
}

func fieldErrors(err error) (map[string]string, string, bool) {
	var formErrs apperrors.FieldErrors
	if errors.As(err, &formErrs) {
		return formErrs, "", true
	}
	var validationErr utils.ValidationError
	if errors.As(err, &validationErr) {
		fields := make(map[string]string, len(validationErr.Errors))
		for _, fieldErr := range validationErr.Errors {
			key := fieldKey(fieldErr.Field)
			if _, exists := fields[key]; !exists {
				fields[key] = fieldErr.Message
			}
		}
		return fields, "", true
	}
	var apiErr utils.APIErrorResponse
	if errors.As(err, &apiErr) && apiErr.ErrorCode == http.StatusBadRequest {
		return nil, "Please check the form", true
	}
	return nil, "", false
}

func fieldKey(field string) string {
	if i := strings.IndexByte(field, '['); i >= 0 {
		field = field[:i]
	}
	if strings.HasPrefix(field, "cast_") {
		return "cast"
	}
	return field
}

func updatedAtToken(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func formatVN(t time.Time, layout string) string {
	return utils.FormatVN(t, layout)
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func at(values []string, i int) string {
	if i < len(values) {
		return values[i]
	}
	return ""
}

func (f flashCookie) set(c *echo.Context, kind, message string) {
	value := base64.RawURLEncoding.EncodeToString([]byte(kind + "|" + message))
	c.SetCookie(f.cookie(value, 60))
}

func (f flashCookie) pop(c *echo.Context) Flash {
	cookie, err := c.Cookie(flashCookieName)
	if err != nil || cookie.Value == "" {
		return Flash{}
	}
	c.SetCookie(f.cookie("", -1))
	raw, err := base64.RawURLEncoding.DecodeString(cookie.Value)
	if err != nil {
		return Flash{}
	}
	kind, message, ok := strings.Cut(string(raw), "|")
	if !ok {
		return Flash{}
	}
	return Flash{Kind: kind, Message: message}
}

func (f flashCookie) cookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     flashCookieName,
		Value:    value,
		Path:     middleware.AdminPathPrefix,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   f.secure,
		SameSite: http.SameSiteLaxMode,
	}
}
