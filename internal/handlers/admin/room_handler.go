package admin

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"cinema-booking/internal/dto"
	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/middleware"
	"cinema-booking/internal/models"
	"cinema-booking/internal/services"
	"cinema-booking/internal/utils"
)

const roomsPath = middleware.AdminPathPrefix + "/rooms"

type RoomRow struct {
	ID          int64
	Name        string
	IsActive    bool
	HasSeats    bool
	TotalSeats  int64
	ActiveSeats int64
}

type RoomListView struct {
	Title       string
	AdminEmail  string
	CSRFToken   string
	Flash       Flash
	TheaterID   int64
	TheaterName string
	Rooms       []RoomRow
}

type RoomFormView struct {
	Title       string
	AdminEmail  string
	CSRFToken   string
	Action      string
	Submit      string
	TheaterID   int64
	TheaterName string
	IsEdit      bool
	Form        dto.AdminRoomForm
	Errors      map[string]string
	Error       string
}

type RoomHandler struct {
	service services.AdminRoomService
	flash   flashCookie
}

func NewRoomHandler(service services.AdminRoomService, secureCookies bool) *RoomHandler {
	return &RoomHandler{service: service, flash: flashCookie{secure: secureCookies}}
}

func (h *RoomHandler) List(c *echo.Context) error {
	theaterID, err := parseID(c)
	if err != nil {
		return err
	}
	theater, rooms, err := h.service.ListByTheater(c.Request().Context(), theaterID)
	if err != nil {
		return utils.ServiceError(err)
	}
	return c.Render(http.StatusOK, "admin/rooms/list", RoomListView{
		Title:       "Rooms",
		AdminEmail:  adminEmail(c),
		CSRFToken:   csrfToken(c),
		Flash:       h.flash.pop(c),
		TheaterID:   theater.ID,
		TheaterName: theater.Name,
		Rooms:       toRoomRows(rooms),
	})
}

func (h *RoomHandler) New(c *echo.Context) error {
	theater, err := h.theater(c)
	if err != nil {
		return err
	}
	return h.renderForm(c, http.StatusOK, h.newView(c, theater, dto.AdminRoomForm{IsActive: true}))
}

func (h *RoomHandler) Create(c *echo.Context) error {
	theater, err := h.theater(c)
	if err != nil {
		return err
	}
	var form dto.AdminRoomForm
	if err := utils.BindAndValidate(c, &form); err != nil {
		return h.renderFormError(c, h.newView(c, theater, form), err)
	}
	if _, err := h.service.Create(c.Request().Context(), theater.ID, form); err != nil {
		return h.renderFormError(c, h.newView(c, theater, form), err)
	}
	h.flash.set(c, flashSuccess, "Room created")
	return c.Redirect(http.StatusSeeOther, roomListPath(theater.ID))
}

func (h *RoomHandler) Edit(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	room, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		return utils.ServiceError(err)
	}
	return h.renderForm(c, http.StatusOK, h.editView(c, room, formFromRoom(room)))
}

func (h *RoomHandler) Update(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	room, err := h.service.Get(ctx, id)
	if err != nil {
		return utils.ServiceError(err)
	}
	var form dto.AdminRoomForm
	if err := utils.BindAndValidate(c, &form); err != nil {
		return h.renderFormError(c, h.editView(c, room, form), err)
	}
	if err := h.service.Update(ctx, id, form); err != nil {
		if errors.Is(err, apperrors.ErrRecordModified) || errors.Is(err, apperrors.ErrRecordTokenInvalid) {
			form.UpdatedAt = updatedAtToken(room.UpdatedAt)
		}
		return h.renderFormError(c, h.editView(c, room, form), err)
	}
	h.flash.set(c, flashSuccess, "Room updated")
	return c.Redirect(http.StatusSeeOther, roomListPath(room.TheaterID))
}

func (h *RoomHandler) ChangeStatus(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	room, err := h.service.Get(ctx, id)
	if err != nil {
		return utils.ServiceError(err)
	}
	active, err := h.service.ChangeStatus(ctx, id)
	if err != nil {
		return utils.ServiceError(err)
	}
	if active {
		h.flash.set(c, flashSuccess, "Room activated")
	} else {
		h.flash.set(c, flashSuccess, "Room deactivated")
	}
	return c.Redirect(http.StatusSeeOther, roomListPath(room.TheaterID))
}

func (h *RoomHandler) theater(c *echo.Context) (*models.Theater, error) {
	theaterID, err := parseID(c)
	if err != nil {
		return nil, err
	}
	theater, err := h.service.GetTheater(c.Request().Context(), theaterID)
	if err != nil {
		return nil, utils.ServiceError(err)
	}
	return theater, nil
}

func (h *RoomHandler) newView(c *echo.Context, theater *models.Theater, form dto.AdminRoomForm) RoomFormView {
	return RoomFormView{
		Title:       "New room",
		AdminEmail:  adminEmail(c),
		CSRFToken:   csrfToken(c),
		Action:      roomListPath(theater.ID),
		Submit:      "Create room",
		TheaterID:   theater.ID,
		TheaterName: theater.Name,
		Form:        form,
	}
}

func (h *RoomHandler) editView(c *echo.Context, room *models.Room, form dto.AdminRoomForm) RoomFormView {
	return RoomFormView{
		Title:       "Edit room",
		AdminEmail:  adminEmail(c),
		CSRFToken:   csrfToken(c),
		Action:      roomsPath + "/" + strconv.FormatInt(room.ID, 10),
		Submit:      "Save changes",
		TheaterID:   room.TheaterID,
		TheaterName: room.Theater.Name,
		IsEdit:      true,
		Form:        form,
	}
}

func (h *RoomHandler) renderForm(c *echo.Context, status int, view RoomFormView) error {
	if view.Errors == nil {
		view.Errors = map[string]string{}
	}
	return c.Render(status, "admin/rooms/form", view)
}

func (h *RoomHandler) renderFormError(c *echo.Context, view RoomFormView, err error) error {
	switch {
	case errors.Is(err, apperrors.ErrRoomNameTaken):
		view.Errors = map[string]string{"name": "Name already exists in this theater"}
	case errors.Is(err, apperrors.ErrRecordModified):
		view.Error = "Someone else changed this room while you were editing. Reload to see the latest data, or save again to overwrite it."
		return h.renderForm(c, http.StatusConflict, view)
	case errors.Is(err, apperrors.ErrRecordTokenInvalid):
		view.Error = "The form was submitted without a valid version token, so nothing was saved. Submit again to retry."
		return h.renderForm(c, http.StatusBadRequest, view)
	default:
		fields, message, ok := fieldErrors(err)
		if !ok {
			return utils.ServiceError(err)
		}
		view.Errors, view.Error = fields, message
	}
	return h.renderForm(c, http.StatusUnprocessableEntity, view)
}

func toRoomRows(rooms []services.RoomWithSeats) []RoomRow {
	rows := make([]RoomRow, 0, len(rooms))
	for _, room := range rooms {
		rows = append(rows, RoomRow{
			ID:          room.ID,
			Name:        room.Name,
			IsActive:    room.IsActive,
			HasSeats:    room.SeatCount.Total > 0,
			TotalSeats:  room.SeatCount.Total,
			ActiveSeats: room.SeatCount.Active,
		})
	}
	return rows
}

func formFromRoom(room *models.Room) dto.AdminRoomForm {
	return dto.AdminRoomForm{
		Name:      room.Name,
		IsActive:  room.IsActive,
		UpdatedAt: updatedAtToken(room.UpdatedAt),
	}
}

func roomListPath(theaterID int64) string {
	return theatersPath + "/" + strconv.FormatInt(theaterID, 10) + "/rooms"
}
