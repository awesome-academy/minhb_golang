package admin

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v5"

	"cinema-booking/internal/realtime"
	"cinema-booking/internal/services"
	"cinema-booking/internal/utils"
)

type NotificationHandler struct {
	service  services.AdminNotificationService
	hub      *realtime.Hub
	upgrader websocket.Upgrader
}

func NewNotificationHandler(service services.AdminNotificationService, hub *realtime.Hub) *NotificationHandler {
	return &NotificationHandler{service: service, hub: hub}
}

func (h *NotificationHandler) Recent(c *echo.Context) error {
	items, err := h.service.Recent(c.Request().Context())
	if err != nil {
		return utils.ServiceError(err)
	}
	return c.JSON(http.StatusOK, items)
}

func (h *NotificationHandler) Socket(c *echo.Context) error {
	conn, err := h.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return nil
	}
	h.hub.Serve(conn)
	return nil
}
