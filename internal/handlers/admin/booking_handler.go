package admin

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/labstack/echo/v5"

	"cinema-booking/internal/dto"
	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/middleware"
	"cinema-booking/internal/models"
	"cinema-booking/internal/services"
	"cinema-booking/internal/utils"
)

const (
	dateTimeLayout = dateLayout + " " + timeLayout

	seatAvailable = "available"
	seatHeld      = "held"
	seatPaid      = "paid"
	seatBlocked   = "blocked"
)

type CounterSeatView struct {
	ID        int64
	Number    int16
	TypeCode  string
	Status    string
	Price     string
	BookingID int64
	PairID    int64
	Title     string
}

type CounterRowView struct {
	Label string
	Seats []CounterSeatView
}

type PriceLegend struct {
	Code  string
	Name  string
	Price string
}

type BookingView struct {
	ID          int64
	Pending     bool
	Code        string
	Status      string
	Customer    string
	Email       string
	Seats       string
	Total       string
	ExpiresAt   string
	ConfirmedAt string
	Note        string
}

type CounterSeatMapView struct {
	Title      string
	AdminEmail string
	CSRFToken  string
	Flash      Flash
	ShowtimeID int64
	Movie      string
	Theater    string
	Room       string
	Date       string
	Time       string
	Format     string
	Status     string
	Badge      string
	Published  bool
	CanSell    bool
	Prices     []PriceLegend
	Rows       []CounterRowView
	Bookings   []BookingView
	Available  int
	Held       int
	Paid       int
	SellURL    string
}

type BookingHandler struct {
	service services.AdminBookingService
	flash   flashCookie
}

func NewBookingHandler(service services.AdminBookingService, secureCookies bool) *BookingHandler {
	return &BookingHandler{service: service, flash: flashCookie{secure: secureCookies}}
}

func (h *BookingHandler) SeatMap(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	data, err := h.service.SeatMap(c.Request().Context(), id)
	if err != nil {
		return utils.ServiceError(err)
	}
	view := buildCounterSeatMapView(data, time.Now())
	view.AdminEmail = adminEmail(c)
	view.CSRFToken = csrfToken(c)
	view.Flash = h.flash.pop(c)
	return c.Render(http.StatusOK, "admin/bookings/seat-map", view)
}

func (h *BookingHandler) Sell(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var form dto.AdminCounterSaleForm
	if err := utils.BindAndValidate(c, &form); err != nil {
		h.flash.set(c, flashDanger, formMessage(err))
		return c.Redirect(http.StatusSeeOther, showtimeSeatsPath(id))
	}
	booking, err := h.service.Sell(c.Request().Context(), id, middleware.AdminUser(c).ID, form)
	switch {
	case errors.Is(err, apperrors.ErrCounterClosed):
		h.flash.set(c, flashDanger, "Counter sales are closed for this showtime")
	case errors.Is(err, apperrors.ErrSeatsInvalid):
		h.flash.set(c, flashDanger, "One or more seats are not available for this showtime")
	case errors.Is(err, apperrors.ErrSeatsTaken):
		h.flash.set(c, flashDanger, "One or more seats have just been taken")
	case errors.Is(err, apperrors.ErrCoupleSeatsUnpaired):
		h.flash.set(c, flashDanger, "Couple seats sell in pairs (1-2, 3-4, ...), select both seats")
	case err != nil:
		return utils.ServiceError(err)
	default:
		h.flash.set(c, flashSuccess, fmt.Sprintf("Sold %d seat(s) at counter · booking %s · total %s",
			len(booking.Tickets), booking.Code, booking.Subtotal.StringFixed(2)))
	}
	return c.Redirect(http.StatusSeeOther, showtimeSeatsPath(id))
}

func (h *BookingHandler) Confirm(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	var form dto.AdminConfirmBookingForm
	if err := utils.BindAndValidate(c, &form); err != nil {
		h.flash.set(c, flashDanger, formMessage(err))
		return c.Redirect(http.StatusSeeOther, confirmBackURL(form.ShowtimeID))
	}
	booking, err := h.service.Confirm(c.Request().Context(), id, form)
	back := confirmBackURL(form.ShowtimeID)
	switch {
	case errors.Is(err, apperrors.ErrBookingCodeMismatch):
		h.flash.set(c, flashDanger, "Booking code does not match")
	case errors.Is(err, apperrors.ErrBookingExpired):
		h.flash.set(c, flashDanger, "Booking hold has expired, its seats can be sold again")
	case errors.Is(err, apperrors.ErrBookingCancelled):
		h.flash.set(c, flashDanger, "Booking was cancelled with its showtime")
	case errors.Is(err, apperrors.ErrBookingAlreadyConfirmed):
		h.flash.set(c, flashSuccess, "Booking is already paid, nothing changed")
	case err != nil:
		return utils.ServiceError(err)
	default:
		h.flash.set(c, flashSuccess, "Booking "+booking.Code+" confirmed as paid")
		back = showtimeSeatsPath(booking.ShowtimeID)
	}
	return c.Redirect(http.StatusSeeOther, back)
}

func buildCounterSeatMapView(data *services.CounterSeatMap, now time.Time) CounterSeatMapView {
	view := counterHeaderView(data.Showtime, now)
	prices := make(map[int64]string, len(data.Showtime.Prices))
	for _, price := range data.Showtime.Prices {
		prices[price.SeatTypeID] = price.Price.StringFixed(2)
	}
	bookingBySeat := make(map[int64]*models.Booking)
	for i := range data.Bookings {
		for _, ticket := range data.Bookings[i].Tickets {
			bookingBySeat[ticket.SeatID] = &data.Bookings[i]
		}
	}
	seatsByBooking := make(map[int64][]string, len(data.Bookings))
	seenTypes := make(map[int64]bool)
	for _, seat := range data.Seats {
		label := seat.RowLabel + strconv.Itoa(int(seat.SeatNumber))
		booking := bookingBySeat[seat.ID]
		if booking != nil {
			seatsByBooking[booking.ID] = append(seatsByBooking[booking.ID], label)
		}
		if !seenTypes[seat.SeatTypeID] {
			seenTypes[seat.SeatTypeID] = true
			view.Prices = append(view.Prices, PriceLegend{Code: seat.SeatType.Code, Name: seat.SeatType.Name, Price: prices[seat.SeatTypeID]})
		}
		view.addSeat(seat.RowLabel, counterSeatView(seat, label, prices[seat.SeatTypeID], booking))
	}
	view.linkCouplePairs()
	view.Bookings = bookingViews(data.Bookings, seatsByBooking)
	return view
}

func (v *CounterSeatMapView) linkCouplePairs() {
	for r := range v.Rows {
		seats := v.Rows[r].Seats
		byNumber := make(map[int16]*CounterSeatView, len(seats))
		for i := range seats {
			byNumber[seats[i].Number] = &seats[i]
		}
		for i := range seats {
			seat := &seats[i]
			if seat.TypeCode != models.SeatTypeCodeCouple || seat.Status != seatAvailable {
				continue
			}
			if partner := byNumber[models.CoupleSeatPartner(seat.Number)]; partner != nil && partner.Status == seatAvailable {
				seat.PairID = partner.ID
			}
		}
	}
}

func counterHeaderView(showtime *models.Showtime, now time.Time) CounterSeatMapView {
	status := showtimeStatusOf(showtime.State(now))
	return CounterSeatMapView{
		Title:      "Seats · " + showtime.Movie.Title,
		ShowtimeID: showtime.ID,
		Movie:      showtime.Movie.Title,
		Theater:    showtime.Room.Theater.Name,
		Room:       showtime.Room.Name,
		Date:       formatVN(showtime.StartsAt, dateLayout),
		Time:       formatVN(showtime.StartsAt, timeLayout) + "–" + formatVN(showtime.EndsAt, timeLayout),
		Format:     string(showtime.Format),
		Status:     status.Label,
		Badge:      status.Badge,
		Published:  showtime.IsPublished,
		CanSell:    showtime.CounterOpen(now),
		SellURL:    showtimesPath + "/" + strconv.FormatInt(showtime.ID, 10) + "/counter-sales",
	}
}

func (v *CounterSeatMapView) addSeat(rowLabel string, seat CounterSeatView) {
	switch seat.Status {
	case seatAvailable:
		v.Available++
	case seatHeld:
		v.Held++
	case seatPaid:
		v.Paid++
	}
	if n := len(v.Rows); n == 0 || v.Rows[n-1].Label != rowLabel {
		v.Rows = append(v.Rows, CounterRowView{Label: rowLabel})
	}
	row := &v.Rows[len(v.Rows)-1]
	row.Seats = append(row.Seats, seat)
}

func counterSeatView(seat models.Seat, label, price string, booking *models.Booking) CounterSeatView {
	view := CounterSeatView{ID: seat.ID, Number: seat.SeatNumber, TypeCode: seat.SeatType.Code, Price: price}
	switch {
	case booking != nil && booking.Status == models.BookingStatusConfirmed:
		view.Status, view.BookingID = seatPaid, booking.ID
		view.Title = label + " · paid · " + booking.Code
	case booking != nil:
		view.Status, view.BookingID = seatHeld, booking.ID
		view.Title = label + " · held · " + booking.User.Email
		if booking.ExpiresAt != nil {
			view.Title += " · until " + formatVN(*booking.ExpiresAt, timeLayout)
		}
	case !seat.IsActive:
		view.Status, view.Title = seatBlocked, label+" · disabled"
	case price == "":
		view.Status, view.Title = seatBlocked, label+" · "+seat.SeatType.Code+" · no price"
	default:
		view.Status, view.Title = seatAvailable, label+" · "+seat.SeatType.Code+" · "+price
	}
	return view
}

func bookingViews(bookings []models.Booking, seatsByBooking map[int64][]string) []BookingView {
	views := make([]BookingView, 0, len(bookings))
	for i := range bookings {
		booking := &bookings[i]
		view := BookingView{
			ID:       booking.ID,
			Pending:  booking.Status == models.BookingStatusPending,
			Status:   string(booking.Status),
			Customer: booking.User.FullName,
			Email:    booking.User.Email,
			Seats:    strings.Join(seatsByBooking[booking.ID], ", "),
			Total:    booking.Subtotal.Sub(booking.DiscountAmount).StringFixed(2) + " " + booking.Currency,
			Note:     derefString(booking.Note),
		}
		if booking.ExpiresAt != nil {
			view.ExpiresAt = formatVN(*booking.ExpiresAt, dateTimeLayout)
		}
		if booking.ConfirmedAt != nil {
			view.ConfirmedAt = formatVN(*booking.ConfirmedAt, dateTimeLayout)
		}
		if booking.Status == models.BookingStatusConfirmed {
			view.Code = booking.Code
		}
		views = append(views, view)
	}
	return views
}

func formMessage(err error) string {
	var validationErr utils.ValidationError
	if errors.As(err, &validationErr) {
		return validationErr.ErrorMessage
	}
	return "Please check the form"
}

func showtimeSeatsPath(showtimeID int64) string {
	return showtimesPath + "/" + strconv.FormatInt(showtimeID, 10) + "/seats"
}

func confirmBackURL(showtimeID int64) string {
	if showtimeID <= 0 {
		return showtimesPath
	}
	return showtimeSeatsPath(showtimeID)
}
