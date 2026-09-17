package admin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/shopspring/decimal"

	"cinema-booking/internal/dto"
	apperrors "cinema-booking/internal/errors"
	"cinema-booking/internal/middleware"
	"cinema-booking/internal/models"
	"cinema-booking/internal/repositories"
	"cinema-booking/internal/services"
	"cinema-booking/internal/utils"
)

const (
	showtimesPath = middleware.AdminPathPrefix + "/showtimes"
	timeLayout    = "15:04"
)

var (
	showtimeFormats = []string{
		string(models.ShowtimeFormat2D), string(models.ShowtimeFormat3D), string(models.ShowtimeFormatIMAX),
	}
	showtimeStatuses = []ShowtimeStatus{
		{Value: string(models.ShowtimeStateScheduled), Label: "Scheduled", Badge: "success"},
		{Value: string(models.ShowtimeStatePlaying), Label: "Playing", Badge: "primary"},
		{Value: string(models.ShowtimeStateFinished), Label: "Finished", Badge: "secondary"},
		{Value: string(models.ShowtimeStateCancelled), Label: "Cancelled", Badge: "danger"},
	}
)

type ShowtimeStatus struct {
	Value string
	Label string
	Badge string
}

type ShowtimeRow struct {
	ID        int64
	Date      string
	Time      string
	Theater   string
	Movie     string
	Room      string
	Format    string
	PriceFrom string
	Status    string
	Badge     string
	Bookings  int64
	Published bool
	CanEdit   bool
	EditURL   string
}

type TheaterOption struct {
	ID   int64
	Name string
}

type ShowtimeListView struct {
	Title           string
	AdminEmail      string
	CSRFToken       string
	Flash           Flash
	Theaters        []TheaterOption
	TheaterID       int64
	From            string
	To              string
	Status          string
	Statuses        []ShowtimeStatus
	PublishedFilter string
	HasFilter       bool
	Back            string
	Rows            []ShowtimeRow
	NewURL          string
	Total           int64
	Page            int
	TotalPages      int
	PrevURL         string
	NextURL         string
	FirstURL        string
	HasPrev         bool
	HasNext         bool
}

type showtimeQuery struct {
	TheaterID int64
	From      string
	To        string
	Status    string
	Published string
}

type MovieOption struct {
	ID          int64
	Title       string
	DurationMin int
}

type RoomOption struct {
	ID   int64
	Name string
}

type RoomGroup struct {
	TheaterID int64
	Theater   string
	Rooms     []RoomOption
}

type PriceRow struct {
	SeatTypeID int64
	Name       string
	Value      string
	Error      string
}

type ShowtimeFormView struct {
	Title         string
	AdminEmail    string
	CSRFToken     string
	Action        string
	Submit        string
	BackURL       string
	IsEdit        bool
	Locked        bool
	PublishLocked bool
	Movies        []MovieOption
	RoomGroups    []RoomGroup
	Formats       []string
	Prices        []PriceRow
	MovieEndsAt   string
	Back          string
	Form          dto.AdminShowtimeForm
	Errors        map[string]string
	Error         string
}

type ShowtimeHandler struct {
	service services.AdminShowtimeService
	flash   flashCookie
}

func NewShowtimeHandler(service services.AdminShowtimeService, secureCookies bool) *ShowtimeHandler {
	return &ShowtimeHandler{service: service, flash: flashCookie{secure: secureCookies}}
}

func (h *ShowtimeHandler) List(c *echo.Context) error {
	filter, query := parseShowtimeValues(c.QueryParams())
	page := parsePage(c.QueryParam("page"))
	back := query.encodePage(page)
	list, err := h.service.List(c.Request().Context(), filter, page)
	if err != nil {
		return utils.ServiceError(err)
	}
	totalPages := int((list.Total + services.ShowtimePageSize - 1) / services.ShowtimePageSize)
	if totalPages < 1 {
		totalPages = 1
	}
	return c.Render(http.StatusOK, "admin/showtimes/list", ShowtimeListView{
		Title:           "Showtimes",
		AdminEmail:      adminEmail(c),
		CSRFToken:       csrfToken(c),
		Flash:           h.flash.pop(c),
		Theaters:        theaterOptions(list.Theaters),
		TheaterID:       query.TheaterID,
		From:            query.From,
		To:              query.To,
		Status:          query.Status,
		Statuses:        showtimeStatuses,
		PublishedFilter: query.Published,
		HasFilter:       query.encode() != "",
		Back:            back,
		Rows:            toShowtimeRows(list.Showtimes, list.BookingCounts, back, time.Now()),
		NewURL:          withQuery(showtimesPath+"/new", back),
		Total:           list.Total,
		Page:            page,
		TotalPages:      totalPages,
		PrevURL:         query.pageURL(page - 1),
		NextURL:         query.pageURL(page + 1),
		FirstURL:        query.pageURL(1),
		HasPrev:         page > 1,
		HasNext:         page < totalPages,
	})
}

func (h *ShowtimeHandler) New(c *echo.Context) error {
	_, query := parseShowtimeValues(c.QueryParams())
	form := dto.AdminShowtimeForm{Format: string(models.ShowtimeFormat2D)}
	view, err := h.formView(c, query.TheaterID, form, nil)
	if err != nil {
		return utils.ServiceError(err)
	}
	return c.Render(http.StatusOK, "admin/showtimes/form", view)
}

func (h *ShowtimeHandler) Create(c *echo.Context) error {
	var form dto.AdminShowtimeForm
	if err := utils.BindAndValidate(c, &form); err != nil {
		return h.renderFormError(c, form, nil, err)
	}
	if _, err := h.service.Create(c.Request().Context(), form); err != nil {
		return h.renderFormError(c, form, nil, err)
	}
	h.flash.set(c, flashSuccess, "Showtime created")
	return c.Redirect(http.StatusSeeOther, listURL(c))
}

func (h *ShowtimeHandler) Edit(c *echo.Context) error {
	detail, err := h.editableDetail(c)
	if err != nil || detail == nil {
		return err
	}
	view, err := h.formView(c, 0, formFromShowtime(detail.Showtime), detail)
	if err != nil {
		return utils.ServiceError(err)
	}
	return c.Render(http.StatusOK, "admin/showtimes/form", view)
}

func (h *ShowtimeHandler) Update(c *echo.Context) error {
	detail, err := h.editableDetail(c)
	if err != nil || detail == nil {
		return err
	}
	var form dto.AdminShowtimeForm
	if err := utils.BindAndValidate(c, &form); err != nil {
		return h.renderFormError(c, form, detail, err)
	}
	_, err = h.service.Update(c.Request().Context(), detail.Showtime.ID, form)
	switch {
	case errors.Is(err, apperrors.ErrShowtimeNotEditable):
		return h.redirectNotEditable(c, "This showtime can no longer be edited")
	case errors.Is(err, apperrors.ErrRecordModified), errors.Is(err, apperrors.ErrRecordTokenInvalid):
		form.UpdatedAt = updatedAtToken(detail.Showtime.UpdatedAt)
		return h.renderFormError(c, form, detail, err)
	case err != nil:
		return h.renderFormError(c, form, detail, err)
	}
	h.flash.set(c, flashSuccess, "Showtime updated")
	return c.Redirect(http.StatusSeeOther, listURL(c))
}

func (h *ShowtimeHandler) ChangePublished(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	published, err := h.service.ChangePublished(c.Request().Context(), id)
	switch {
	case errors.Is(err, apperrors.ErrShowtimeHasBookings):
		h.flash.set(c, flashDanger, "Cannot unpublish: this showtime has bookings")
	case errors.Is(err, apperrors.ErrShowtimeNotEditable):
		h.flash.set(c, flashDanger, "This showtime can no longer be changed")
	case err != nil:
		return utils.ServiceError(err)
	case published:
		h.flash.set(c, flashSuccess, "Showtime published")
	default:
		h.flash.set(c, flashSuccess, "Showtime unpublished")
	}
	return c.Redirect(http.StatusSeeOther, listURL(c))
}

func (h *ShowtimeHandler) Cancel(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	cancelled, err := h.service.Cancel(c.Request().Context(), id)
	switch {
	case errors.Is(err, apperrors.ErrShowtimeNotEditable):
		h.flash.set(c, flashDanger, "This showtime can no longer be cancelled")
	case err != nil:
		return utils.ServiceError(err)
	default:
		h.flash.set(c, flashSuccess, cancelFlash(cancelled))
	}
	return c.Redirect(http.StatusSeeOther, listURL(c))
}

func (h *ShowtimeHandler) editableDetail(c *echo.Context) (*services.ShowtimeDetail, error) {
	id, err := parseID(c)
	if err != nil {
		return nil, err
	}
	detail, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		return nil, utils.ServiceError(err)
	}
	if !detail.Showtime.Editable(time.Now()) {
		return nil, h.redirectNotEditable(c, "This showtime can no longer be edited")
	}
	return detail, nil
}

func (h *ShowtimeHandler) redirectNotEditable(c *echo.Context, message string) error {
	h.flash.set(c, flashDanger, message)
	return c.Redirect(http.StatusSeeOther, listURL(c))
}

func (h *ShowtimeHandler) formView(c *echo.Context, theaterID int64, form dto.AdminShowtimeForm, detail *services.ShowtimeDetail) (ShowtimeFormView, error) {
	data, err := h.service.FormData(c.Request().Context(), theaterID)
	if err != nil {
		return ShowtimeFormView{}, err
	}
	view := ShowtimeFormView{
		Title:      "New showtime",
		AdminEmail: adminEmail(c),
		CSRFToken:  csrfToken(c),
		Action:     showtimesPath,
		Submit:     "Create showtime",
		BackURL:    listURL(c),
		Formats:    showtimeFormats,
		Back:       backQuery(c),
		Form:       form,
		Errors:     map[string]string{},
	}
	var current *models.Showtime
	if detail != nil {
		current = detail.Showtime
		view.Title = "Edit showtime"
		view.Action = showtimesPath + "/" + strconv.FormatInt(current.ID, 10)
		view.Submit = "Save changes"
		view.IsEdit = true
		view.Locked = detail.HasBookings
		view.PublishLocked = detail.HasBookings && current.IsPublished
		if view.PublishLocked {
			view.Form.IsPublished = true
		}
	}
	view.Movies = movieOptions(data.Movies, current)
	view.RoomGroups = roomGroups(data.Rooms, current)
	if view.Form.RoomID == 0 {
		view.Form.RoomID = firstRoomID(data.Rooms, theaterID)
	}
	view.Prices = priceRows(data.SeatTypes, form, data.Prefill)
	view.MovieEndsAt = movieEndsAt(view.Movies, form)
	return view, nil
}

func (h *ShowtimeHandler) renderFormError(c *echo.Context, form dto.AdminShowtimeForm, detail *services.ShowtimeDetail, err error) error {
	status := http.StatusUnprocessableEntity
	var fields map[string]string
	var message string
	var overlap *apperrors.ShowtimeOverlapError
	switch {
	case errors.Is(err, apperrors.ErrShowtimeMovieInvalid):
		fields = map[string]string{"movie_id": "Movie is not available"}
	case errors.Is(err, apperrors.ErrShowtimeRoomInvalid):
		fields = map[string]string{"room_id": "Room is not available"}
	case errors.Is(err, apperrors.ErrRoomHasNoSeats):
		fields = map[string]string{"room_id": "Room has no seats yet. Generate its seat map first"}
	case errors.Is(err, apperrors.ErrShowtimeInPast):
		fields = map[string]string{"starts_at": "Starts at must be in the future"}
	case errors.Is(err, apperrors.ErrShowtimeEndsTooEarly):
		fields = map[string]string{"ends_at": "Ends at must not be before the movie ends"}
	case errors.As(err, &overlap):
		fields = map[string]string{"starts_at": fmt.Sprintf(`Overlaps with "%s" (%s–%s) in this room`,
			overlap.MovieTitle, formatVN(overlap.StartsAt, timeLayout), formatVN(overlap.EndsAt, timeLayout))}
	case errors.Is(err, apperrors.ErrShowtimeLocked):
		message = "Movie, room and time cannot be changed because this showtime has bookings"
	case errors.Is(err, apperrors.ErrShowtimeHasBookings):
		message = "Cannot unpublish: this showtime has bookings"
	case errors.Is(err, apperrors.ErrRecordModified):
		message = "Someone else changed this showtime while you were editing. Reload to see the latest data, or save again to overwrite it."
		status = http.StatusConflict
	case errors.Is(err, apperrors.ErrRecordTokenInvalid):
		message = "The form was submitted without a valid version token, so nothing was saved. Submit again to retry."
		status = http.StatusBadRequest
	default:
		var ok bool
		if fields, message, ok = fieldErrors(err); !ok {
			return utils.ServiceError(err)
		}
	}
	view, viewErr := h.formView(c, 0, form, detail)
	if viewErr != nil {
		return utils.ServiceError(viewErr)
	}
	if fields != nil {
		view.Errors = fields
	}
	view.Error = message
	for i := range view.Prices {
		view.Prices[i].Error = view.Errors["price_"+strconv.FormatInt(view.Prices[i].SeatTypeID, 10)]
	}
	return c.Render(status, "admin/showtimes/form", view)
}

func parseShowtimeValues(values url.Values) (repositories.ShowtimeFilter, showtimeQuery) {
	theaterID, _ := strconv.ParseInt(values.Get("theater_id"), 10, 64)
	from, fromOK := parseDate(values.Get("from"))
	to, toOK := parseDate(values.Get("to"))
	if fromOK && toOK && to.Before(from) {
		from, to = to, from
	}
	status := values.Get("status")
	if showtimeStatusOf(models.ShowtimeState(status)).Value == "" {
		status = ""
	}
	published := publishedFilter(values.Get("published"))
	filter := repositories.ShowtimeFilter{TheaterID: max(theaterID, 0), State: models.ShowtimeState(status), Published: published}
	query := showtimeQuery{TheaterID: filter.TheaterID, Status: status}
	if published != nil {
		query.Published = values.Get("published")
	}
	if fromOK {
		filter.From, query.From = from, from.Format(dateLayout)
	}
	if toOK {
		filter.To, query.To = to.AddDate(0, 0, 1), to.Format(dateLayout)
	}
	return filter, query
}

func (q showtimeQuery) encode() string {
	return q.values().Encode()
}

func (q showtimeQuery) encodePage(page int) string {
	values := q.values()
	if page > 1 {
		values.Set("page", strconv.Itoa(page))
	}
	return values.Encode()
}

func (q showtimeQuery) values() url.Values {
	values := url.Values{}
	if q.TheaterID > 0 {
		values.Set("theater_id", strconv.FormatInt(q.TheaterID, 10))
	}
	if q.From != "" {
		values.Set("from", q.From)
	}
	if q.To != "" {
		values.Set("to", q.To)
	}
	if q.Status != "" {
		values.Set("status", q.Status)
	}
	if q.Published != "" {
		values.Set("published", q.Published)
	}
	return values
}

func publishedFilter(raw string) *bool {
	switch raw {
	case "published":
		published := true
		return &published
	case "draft":
		published := false
		return &published
	}
	return nil
}

func (q showtimeQuery) pageURL(page int) string {
	return withQuery(showtimesPath, q.encodePage(page))
}

func backQuery(c *echo.Context) string {
	values := c.QueryParams()
	if raw := c.FormValue("back"); raw != "" {
		values, _ = url.ParseQuery(raw)
	}
	_, query := parseShowtimeValues(values)
	return query.encodePage(parsePage(values.Get("page")))
}

func listURL(c *echo.Context) string {
	return withQuery(showtimesPath, backQuery(c))
}

func parseDate(raw string) (time.Time, bool) {
	date, err := time.ParseInLocation(dateLayout, raw, utils.Location)
	return date, err == nil
}

func withQuery(path, query string) string {
	if query == "" {
		return path
	}
	return path + "?" + query
}

func showtimeStatusOf(state models.ShowtimeState) ShowtimeStatus {
	for _, status := range showtimeStatuses {
		if status.Value == string(state) {
			return status
		}
	}
	return ShowtimeStatus{}
}

func theaterOptions(theaters []models.Theater) []TheaterOption {
	options := make([]TheaterOption, 0, len(theaters))
	for _, theater := range theaters {
		options = append(options, TheaterOption{ID: theater.ID, Name: theater.Name})
	}
	return options
}

func cancelFlash(cancelled int64) string {
	switch cancelled {
	case 0:
		return "Showtime cancelled"
	case 1:
		return "Showtime cancelled, 1 booking cancelled and seats released"
	default:
		return fmt.Sprintf("Showtime cancelled, %d bookings cancelled and seats released", cancelled)
	}
}

func toShowtimeRows(showtimes []models.Showtime, counts map[int64]int64, back string, now time.Time) []ShowtimeRow {
	rows := make([]ShowtimeRow, 0, len(showtimes))
	for i := range showtimes {
		showtime := &showtimes[i]
		status := showtimeStatusOf(showtime.State(now))
		rows = append(rows, ShowtimeRow{
			ID:        showtime.ID,
			Date:      formatVN(showtime.StartsAt, dateLayout),
			Time:      formatVN(showtime.StartsAt, timeLayout) + "–" + formatVN(showtime.EndsAt, timeLayout),
			Theater:   showtime.Room.Theater.Name,
			Movie:     showtime.Movie.Title,
			Room:      showtime.Room.Name,
			Format:    string(showtime.Format),
			PriceFrom: priceFrom(showtime.Prices),
			Status:    status.Label,
			Badge:     status.Badge,
			Bookings:  counts[showtime.ID],
			Published: showtime.IsPublished,
			CanEdit:   showtime.Editable(now),
			EditURL:   withQuery(showtimesPath+"/"+strconv.FormatInt(showtime.ID, 10)+"/edit", back),
		})
	}
	return rows
}

func priceFrom(prices []models.ShowtimePrice) string {
	if len(prices) == 0 {
		return "—"
	}
	lowest := prices[0].Price
	for _, price := range prices[1:] {
		lowest = decimal.Min(lowest, price.Price)
	}
	return lowest.StringFixed(2)
}

func formFromShowtime(showtime *models.Showtime) dto.AdminShowtimeForm {
	form := dto.AdminShowtimeForm{
		MovieID:     showtime.MovieID,
		RoomID:      showtime.RoomID,
		StartsAt:    formatVN(showtime.StartsAt, services.ShowtimeInputLayout),
		EndsAt:      formatVN(showtime.EndsAt, services.ShowtimeInputLayout),
		Format:      string(showtime.Format),
		IsPublished: showtime.IsPublished,
		UpdatedAt:   updatedAtToken(showtime.UpdatedAt),
	}
	for _, price := range showtime.Prices {
		form.SeatTypeIDs = append(form.SeatTypeIDs, price.SeatTypeID)
		form.Prices = append(form.Prices, price.Price.StringFixed(2))
	}
	return form
}

func movieOptions(movies []models.Movie, current *models.Showtime) []MovieOption {
	options := make([]MovieOption, 0, len(movies)+1)
	found := false
	for _, movie := range movies {
		found = found || (current != nil && movie.ID == current.MovieID)
		options = append(options, MovieOption{ID: movie.ID, Title: movie.Title, DurationMin: movie.DurationMin})
	}
	if current != nil && !found {
		options = append(options, MovieOption{ID: current.MovieID, Title: current.Movie.Title, DurationMin: current.Movie.DurationMin})
	}
	return options
}

func roomGroups(rooms []models.Room, current *models.Showtime) []RoomGroup {
	var groups []RoomGroup
	found := false
	for _, room := range rooms {
		found = found || (current != nil && room.ID == current.RoomID)
		if n := len(groups); n == 0 || groups[n-1].TheaterID != room.TheaterID {
			groups = append(groups, RoomGroup{TheaterID: room.TheaterID, Theater: room.Theater.Name})
		}
		group := &groups[len(groups)-1]
		group.Rooms = append(group.Rooms, RoomOption{ID: room.ID, Name: room.Name})
	}
	if current != nil && !found {
		groups = append(groups, RoomGroup{
			TheaterID: current.Room.TheaterID,
			Theater:   current.Room.Theater.Name,
			Rooms:     []RoomOption{{ID: current.RoomID, Name: current.Room.Name}},
		})
	}
	return groups
}

func firstRoomID(rooms []models.Room, theaterID int64) int64 {
	for _, room := range rooms {
		if room.TheaterID == theaterID {
			return room.ID
		}
	}
	return 0
}

func priceRows(seatTypes []models.SeatType, form dto.AdminShowtimeForm, prefill map[int64]decimal.Decimal) []PriceRow {
	values := make(map[int64]string, len(seatTypes))
	if len(form.SeatTypeIDs) == 0 {
		for seatTypeID, price := range prefill {
			values[seatTypeID] = price.StringFixed(2)
		}
	}
	for i, seatTypeID := range form.SeatTypeIDs {
		values[seatTypeID] = at(form.Prices, i)
	}
	rows := make([]PriceRow, 0, len(seatTypes))
	for _, seatType := range seatTypes {
		rows = append(rows, PriceRow{SeatTypeID: seatType.ID, Name: seatType.Name, Value: values[seatType.ID]})
	}
	return rows
}

func movieEndsAt(movies []MovieOption, form dto.AdminShowtimeForm) string {
	startsAt, err := time.ParseInLocation(services.ShowtimeInputLayout, form.StartsAt, utils.Location)
	if err != nil {
		return ""
	}
	for _, movie := range movies {
		if movie.ID != form.MovieID {
			continue
		}
		endsAt := startsAt.Add(time.Duration(movie.DurationMin) * time.Minute)
		if endsAt.Day() != startsAt.Day() {
			return endsAt.Format(timeLayout) + " (+1 day)"
		}
		return endsAt.Format(timeLayout)
	}
	return ""
}
