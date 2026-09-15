package admin

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"

	"cinema-booking/internal/dto"
	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/models"
	"cinema-booking/internal/services"
	"cinema-booking/internal/utils"
)

type SeatView struct {
	ID       int64
	Number   int16
	TypeCode string
	IsActive bool
}

type SeatRowView struct {
	Label      string
	SeatTypeID int64
	Seats      []SeatView
}

type SeatMapView struct {
	Title        string
	AdminEmail   string
	CSRFToken    string
	Flash        Flash
	RoomID       int64
	RoomName     string
	TheaterID    int64
	TheaterName  string
	Rows         []SeatRowView
	SeatTypes    []models.SeatType
	TotalSeats   int
	HasSeats     bool
	HasShowtimes bool
	Form         dto.AdminSeatGenerateForm
	Errors       map[string]string
	Error        string
}

type SeatHandler struct {
	service services.AdminSeatService
	flash   flashCookie
}

func NewSeatHandler(service services.AdminSeatService, secureCookies bool) *SeatHandler {
	return &SeatHandler{service: service, flash: flashCookie{secure: secureCookies}}
}

func (h *SeatHandler) Map(c *echo.Context) error {
	roomID, err := parseID(c)
	if err != nil {
		return err
	}
	view, err := h.mapView(c, roomID)
	if err != nil {
		return utils.ServiceError(err)
	}
	view.Flash = h.flash.pop(c)
	view.Form = defaultGenerateForm(view)
	return c.Render(http.StatusOK, "admin/seats/map", view)
}

func (h *SeatHandler) Generate(c *echo.Context) error {
	roomID, err := parseID(c)
	if err != nil {
		return err
	}
	var form dto.AdminSeatGenerateForm
	if err := utils.BindAndValidate(c, &form); err != nil {
		return h.renderMapError(c, roomID, form, err)
	}
	count, err := h.service.Generate(c.Request().Context(), roomID, form)
	switch {
	case errors.Is(err, apperrors.ErrRoomHasShowtimes):
		h.flash.set(c, flashDanger, "Cannot generate: the seat map is locked because this room has showtimes")
	case err != nil:
		return h.renderMapError(c, roomID, form, err)
	default:
		h.flash.set(c, flashSuccess, fmt.Sprintf("Generated %d seats", count))
	}
	return c.Redirect(http.StatusSeeOther, seatMapPath(roomID))
}

func (h *SeatHandler) ChangeRowTypes(c *echo.Context) error {
	roomID, err := parseID(c)
	if err != nil {
		return err
	}
	var form dto.AdminSeatRowTypesForm
	if err := utils.BindAndValidate(c, &form); err != nil || len(form.RowLabels) != len(form.SeatTypeIDs) {
		h.flash.set(c, flashDanger, "Please check the row seat types")
		return c.Redirect(http.StatusSeeOther, seatMapPath(roomID))
	}
	changed, err := h.service.ChangeRowTypes(c.Request().Context(), roomID, form)
	switch {
	case errors.Is(err, apperrors.ErrSeatTypeInvalid):
		h.flash.set(c, flashDanger, "Seat type does not exist")
	case err != nil:
		return utils.ServiceError(err)
	case changed == 0:
		h.flash.set(c, flashSuccess, "No seat type changes")
	case changed == 1:
		h.flash.set(c, flashSuccess, "Seat type updated for 1 row")
	default:
		h.flash.set(c, flashSuccess, fmt.Sprintf("Seat types updated for %d rows", changed))
	}
	return c.Redirect(http.StatusSeeOther, seatMapPath(roomID))
}

func (h *SeatHandler) ChangeStatus(c *echo.Context) error {
	roomID, err := parseID(c)
	if err != nil {
		return err
	}
	seatID, err := parseParamID(c, "seatId")
	if err != nil {
		return err
	}
	seat, err := h.service.ChangeStatus(c.Request().Context(), roomID, seatID)
	if err != nil {
		return utils.ServiceError(err)
	}
	return c.JSON(http.StatusOK, dto.AdminSeatStatusResponse{IsActive: seat.IsActive})
}

func (h *SeatHandler) mapView(c *echo.Context, roomID int64) (SeatMapView, error) {
	seatMap, err := h.service.Map(c.Request().Context(), roomID)
	if err != nil {
		return SeatMapView{}, err
	}
	return SeatMapView{
		Title:        "Seats",
		AdminEmail:   adminEmail(c),
		CSRFToken:    csrfToken(c),
		RoomID:       seatMap.Room.ID,
		RoomName:     seatMap.Room.Name,
		TheaterID:    seatMap.Room.TheaterID,
		TheaterName:  seatMap.Room.Theater.Name,
		Rows:         buildSeatRows(seatMap.Seats),
		SeatTypes:    seatMap.SeatTypes,
		TotalSeats:   len(seatMap.Seats),
		HasSeats:     len(seatMap.Seats) > 0,
		HasShowtimes: seatMap.HasShowtimes,
		Errors:       map[string]string{},
	}, nil
}

func (h *SeatHandler) renderMapError(c *echo.Context, roomID int64, form dto.AdminSeatGenerateForm, err error) error {
	fields, message, ok := fieldErrors(err)
	if errors.Is(err, apperrors.ErrSeatTypeInvalid) {
		fields, ok = map[string]string{"seat_type_id": "Seat type does not exist"}, true
	}
	if !ok {
		return utils.ServiceError(err)
	}
	view, viewErr := h.mapView(c, roomID)
	if viewErr != nil {
		return utils.ServiceError(viewErr)
	}
	view.Form, view.Error = form, message
	if fields != nil {
		view.Errors = fields
	}
	return c.Render(http.StatusUnprocessableEntity, "admin/seats/map", view)
}

func buildSeatRows(seats []models.Seat) []SeatRowView {
	var rows []SeatRowView
	for _, seat := range seats {
		if n := len(rows); n == 0 || rows[n-1].Label != seat.RowLabel {
			rows = append(rows, SeatRowView{Label: seat.RowLabel, SeatTypeID: seat.SeatTypeID})
		}
		row := &rows[len(rows)-1]
		row.Seats = append(row.Seats, SeatView{
			ID:       seat.ID,
			Number:   seat.SeatNumber,
			TypeCode: seat.SeatType.Code,
			IsActive: seat.IsActive,
		})
	}
	return rows
}

func defaultGenerateForm(view SeatMapView) dto.AdminSeatGenerateForm {
	form := dto.AdminSeatGenerateForm{Rows: len(view.Rows)}
	if len(view.SeatTypes) > 0 {
		form.SeatTypeID = view.SeatTypes[0].ID
	}
	for _, row := range view.Rows {
		form.SeatsPerRow = max(form.SeatsPerRow, len(row.Seats))
	}
	return form
}

func seatMapPath(roomID int64) string {
	return roomsPath + "/" + strconv.FormatInt(roomID, 10) + "/seats"
}
