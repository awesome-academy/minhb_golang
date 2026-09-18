package user

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	"cinema-booking/internal/dto"
	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/middleware"
	"cinema-booking/internal/services"
	"cinema-booking/internal/utils"
)

type BookingHandler struct {
	bookings services.UserBookingService
}

func NewBookingHandler(bookings services.UserBookingService) *BookingHandler {
	return &BookingHandler{bookings: bookings}
}

// @Summary Create a booking
// @Description Holds 1–8 seats of one published showtime for the current user and returns the booking code to pay at the counter. Seats stay held until `expiresAt` (30 minutes before the showtime starts); unpaid bookings expire automatically and their seats are released. One pending booking per user per showtime. Booking closes 30 minutes before the showtime starts. Prices come from the showtime's price table, never from the request.
// @Tags bookings
// @Accept json
// @Produce json
// @Param request body dto.CreateBookingRequest true "Showtime and seat ids (1–8, unique)"
// @Success 201 {object} dto.BookingResponse
// @Failure 400 {object} utils.ValidationError "validation failed; or booking closed (the showtime starts in 30 minutes or less); or a seat is not bookable (wrong room, disabled, unpriced, unknown) — the last two return {errorCode, errorMessage} without `errors`"
// @Failure 401 {object} utils.APIErrorResponse
// @Failure 404 {object} utils.APIErrorResponse "showtime not found, not published, cancelled or already started"
// @Failure 409 {object} utils.APIErrorResponse "a seat was just taken by someone else, or you already have a pending booking for this showtime"
// @Failure 500 {object} utils.APIErrorResponse
// @Security bearerauth
// @Router /bookings [post]
func (h *BookingHandler) Create(c *echo.Context) error {
	user := middleware.CurrentUser(c)
	if user == nil {
		return utils.APIError(http.StatusUnauthorized, "access token required")
	}
	var request dto.CreateBookingRequest
	if err := utils.BindAndValidate(c, &request); err != nil {
		return err
	}

	booking, err := h.bookings.Create(c.Request().Context(), user.ID, request)
	if err != nil {
		return bookingError(err)
	}

	return c.JSON(http.StatusCreated, dto.NewBookingResponse(booking))
}

func bookingError(err error) error {
	switch {
	case errors.Is(err, apperrors.ErrBookingTooLate), errors.Is(err, apperrors.ErrSeatsInvalid):
		return utils.APIError(http.StatusBadRequest, err.Error())
	case errors.Is(err, apperrors.ErrSeatsTaken), errors.Is(err, apperrors.ErrBookingPendingExists):
		return utils.APIError(http.StatusConflict, err.Error())
	default:
		return utils.ServiceError(err)
	}
}
