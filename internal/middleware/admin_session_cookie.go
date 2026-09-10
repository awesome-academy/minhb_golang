package middleware

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
)

const AdminSessionCookieName = "admin_session"

type AdminSessionCookie struct {
	Secure bool
	TTL    time.Duration
}

func (k AdminSessionCookie) Set(c *echo.Context, id string) {
	c.SetCookie(k.cookie(id, int(k.TTL.Seconds())))
}

func (k AdminSessionCookie) Clear(c *echo.Context) {
	c.SetCookie(k.cookie("", -1))
}

func (k AdminSessionCookie) cookie(value string, maxAge int) *http.Cookie {
	return &http.Cookie{
		Name:     AdminSessionCookieName,
		Value:    value,
		Path:     AdminPathPrefix,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   k.Secure,
		SameSite: http.SameSiteLaxMode,
	}
}
