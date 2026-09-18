package user

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/services"
	"cinema-booking/internal/utils"
)

type ShowtimeHandler struct {
	showtimes services.UserShowtimeService
}

func NewShowtimeHandler(showtimes services.UserShowtimeService) *ShowtimeHandler {
	return &ShowtimeHandler{showtimes: showtimes}
}

// @Summary Showtimes of a movie on a day
// @Description Returns the published, scheduled, not-yet-started showtimes of one movie on one calendar day (Asia/Ho_Chi_Minh), grouped by theater. Times are UTC (RFC 3339, `Z`); `priceFrom` is the lowest seat price. Deleted or unknown movies are 404; a day without showtimes returns an empty `theaters` list.
// @Tags showtimes
// @Produce json
// @Param id path int true "Movie ID"
// @Param date query string false "Day in YYYY-MM-DD (Asia/Ho_Chi_Minh), default today" example(2026-09-19)
// @Success 200 {object} dto.MovieScheduleResponse
// @Failure 400 {object} utils.ValidationError
// @Failure 404 {object} utils.APIErrorResponse "movie not found"
// @Failure 500 {object} utils.APIErrorResponse
// @Router /movies/{id}/showtimes [get]
func (h *ShowtimeHandler) ByMovie(c *echo.Context) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return err
	}
	var query dto.ScheduleQuery
	if err := utils.BindAndValidate(c, &query); err != nil {
		return err
	}

	movie, schedule, err := h.showtimes.ByMovie(c.Request().Context(), id, query.Date)
	if err != nil {
		return utils.ServiceError(err)
	}

	return c.JSON(http.StatusOK, dto.NewMovieScheduleResponse(movie, schedule.Date, schedule.Showtimes))
}

// @Summary Showtimes of a theater on a day
// @Description Returns the published, scheduled, not-yet-started showtimes of one active theater on one calendar day (Asia/Ho_Chi_Minh), grouped by movie. Times are UTC (RFC 3339, `Z`); `priceFrom` is the lowest seat price. Inactive or unknown theaters are 404; a day without showtimes returns an empty `movies` list.
// @Tags showtimes
// @Produce json
// @Param id path int true "Theater ID"
// @Param date query string false "Day in YYYY-MM-DD (Asia/Ho_Chi_Minh), default today" example(2026-09-19)
// @Success 200 {object} dto.TheaterScheduleResponse
// @Failure 400 {object} utils.ValidationError
// @Failure 404 {object} utils.APIErrorResponse "theater not found or inactive"
// @Failure 500 {object} utils.APIErrorResponse
// @Router /theaters/{id}/showtimes [get]
func (h *ShowtimeHandler) ByTheater(c *echo.Context) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return err
	}
	var query dto.ScheduleQuery
	if err := utils.BindAndValidate(c, &query); err != nil {
		return err
	}

	theater, schedule, err := h.showtimes.ByTheater(c.Request().Context(), id, query.Date)
	if err != nil {
		return utils.ServiceError(err)
	}

	return c.JSON(http.StatusOK, dto.NewTheaterScheduleResponse(theater, schedule.Date, schedule.Showtimes))
}

// @Summary Seat map of a showtime
// @Description Returns every seat of the room with its status: `available`, `held` (active hold), `sold` (paid) or `blocked` (seat disabled or its seat type has no price for this showtime, `price` is then null). Only published, scheduled, not-yet-started showtimes in active rooms and theaters are visible; anything else is 404.
// @Tags showtimes
// @Produce json
// @Param id path int true "Showtime ID"
// @Success 200 {object} dto.SeatMapResponse
// @Failure 404 {object} utils.APIErrorResponse "showtime not found, not published, cancelled or already started"
// @Failure 500 {object} utils.APIErrorResponse
// @Router /showtimes/{id}/seats [get]
func (h *ShowtimeHandler) SeatMap(c *echo.Context) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return err
	}

	seatMap, err := h.showtimes.SeatMap(c.Request().Context(), id)
	if err != nil {
		return utils.ServiceError(err)
	}

	return c.JSON(http.StatusOK, dto.NewSeatMapResponse(seatMap.Showtime, seatMap.Seats, seatMap.Statuses))
}
