package admin

import (
	"encoding/base64"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v5"

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
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		return 0, utils.APIError(http.StatusNotFound, "resource not found")
	}
	return id, nil
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
