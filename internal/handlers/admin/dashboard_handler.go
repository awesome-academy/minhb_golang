package admin

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"cinema-booking/internal/middleware"
)

type DashboardView struct {
	Title      string
	AdminEmail string
	CSRFToken  string
}

type DashboardHandler struct{}

func NewDashboardHandler() *DashboardHandler {
	return &DashboardHandler{}
}

func (h *DashboardHandler) Index(c *echo.Context) error {
	user := middleware.AdminUser(c)
	if user == nil {
		return c.Redirect(http.StatusFound, middleware.AdminLoginPath)
	}
	return c.Render(http.StatusOK, "admin/dashboard", DashboardView{
		Title:      "Dashboard",
		AdminEmail: user.Email,
		CSRFToken:  csrfToken(c),
	})
}
