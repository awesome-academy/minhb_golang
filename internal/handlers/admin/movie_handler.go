package admin

import (
	"errors"
	"net/http"
	"slices"
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
	moviesPath = middleware.AdminPathPrefix + "/movies"
	dateLayout = "2006-01-02"
)

var (
	ageRatings = []string{
		string(models.AgeRatingP), string(models.AgeRatingK), string(models.AgeRatingT13),
		string(models.AgeRatingT16), string(models.AgeRatingT18), string(models.AgeRatingC),
	}
	movieStatuses = []string{
		string(models.MovieStatusComingSoon), string(models.MovieStatusNowShowing), string(models.MovieStatusEnded),
	}
)

type MovieRow struct {
	ID          int64
	Title       string
	Slug        string
	AgeRating   string
	ReleaseDate string
	Status      string
	Genres      string
}

type MovieListView struct {
	Title      string
	AdminEmail string
	CSRFToken  string
	Flash      Flash
	Q          string
	Movies     []MovieRow
	Total      int64
	Page       int
	TotalPages int
	PrevPage   int
	NextPage   int
	HasPrev    bool
	HasNext    bool
}

type GenreOption struct {
	ID      int64
	Name    string
	Checked bool
}

type CastRow struct {
	Name string
	Role string
}

type MovieFormView struct {
	Title      string
	AdminEmail string
	CSRFToken  string
	Action     string
	Submit     string
	Slug       string
	IsEdit     bool
	Form       dto.AdminMovieForm
	Errors     map[string]string
	Error      string
	Genres     []GenreOption
	Cast       []CastRow
	AgeRatings []string
	Statuses   []string
}

type MovieHandler struct {
	service services.AdminMovieService
	flash   flashCookie
}

func NewMovieHandler(service services.AdminMovieService, secureCookies bool) *MovieHandler {
	return &MovieHandler{service: service, flash: flashCookie{secure: secureCookies}}
}

func (h *MovieHandler) List(c *echo.Context) error {
	q := strings.TrimSpace(c.QueryParam("q"))
	page := parsePage(c.QueryParam("page"))
	movies, total, err := h.service.List(c.Request().Context(), q, page)
	if err != nil {
		return utils.ServiceError(err)
	}
	totalPages := int((total + services.MoviePageSize - 1) / services.MoviePageSize)
	if totalPages < 1 {
		totalPages = 1
	}
	return c.Render(http.StatusOK, "admin/movies/list", MovieListView{
		Title:      "Movies",
		AdminEmail: adminEmail(c),
		CSRFToken:  csrfToken(c),
		Flash:      h.flash.pop(c),
		Q:          q,
		Movies:     toMovieRows(movies),
		Total:      total,
		Page:       page,
		TotalPages: totalPages,
		PrevPage:   page - 1,
		NextPage:   page + 1,
		HasPrev:    page > 1,
		HasNext:    page < totalPages,
	})
}

func (h *MovieHandler) New(c *echo.Context) error {
	form := dto.AdminMovieForm{Status: string(models.MovieStatusComingSoon), AgeRating: string(models.AgeRatingP)}
	return h.renderForm(c, http.StatusOK, h.newView(c, form))
}

func (h *MovieHandler) Create(c *echo.Context) error {
	var form dto.AdminMovieForm
	if err := utils.BindAndValidate(c, &form); err != nil {
		return h.renderFormError(c, h.newView(c, form), err)
	}
	if _, err := h.service.Create(c.Request().Context(), form); err != nil {
		return h.renderFormError(c, h.newView(c, form), err)
	}
	h.flash.set(c, flashSuccess, "Movie created")
	return c.Redirect(http.StatusSeeOther, moviesPath)
}

func (h *MovieHandler) Edit(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	movie, err := h.service.Get(c.Request().Context(), id)
	if err != nil {
		return utils.ServiceError(err)
	}
	return h.renderForm(c, http.StatusOK, h.editView(c, movie, formFromMovie(movie)))
}

func (h *MovieHandler) Update(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	movie, err := h.service.Get(ctx, id)
	if err != nil {
		return utils.ServiceError(err)
	}
	var form dto.AdminMovieForm
	if err := utils.BindAndValidate(c, &form); err != nil {
		return h.renderFormError(c, h.editView(c, movie, form), err)
	}
	if err := h.service.Update(ctx, id, form); err != nil {
		if errors.Is(err, apperrors.ErrMovieModified) || errors.Is(err, apperrors.ErrMovieTokenInvalid) {
			form.UpdatedAt = updatedAtToken(movie.UpdatedAt)
		}
		return h.renderFormError(c, h.editView(c, movie, form), err)
	}
	h.flash.set(c, flashSuccess, "Movie updated")
	return c.Redirect(http.StatusSeeOther, moviesPath)
}

func (h *MovieHandler) Delete(c *echo.Context) error {
	id, err := parseID(c)
	if err != nil {
		return err
	}
	err = h.service.Delete(c.Request().Context(), id)
	switch {
	case errors.Is(err, apperrors.ErrMovieHasShowtimes):
		h.flash.set(c, flashDanger, "Cannot delete: the movie still has upcoming showtimes")
	case err != nil:
		return utils.ServiceError(err)
	default:
		h.flash.set(c, flashSuccess, "Movie deleted")
	}
	return c.Redirect(http.StatusSeeOther, moviesPath)
}

func (h *MovieHandler) newView(c *echo.Context, form dto.AdminMovieForm) MovieFormView {
	return MovieFormView{
		Title:      "New movie",
		AdminEmail: adminEmail(c),
		CSRFToken:  csrfToken(c),
		Action:     moviesPath,
		Submit:     "Create movie",
		Form:       form,
	}
}

func (h *MovieHandler) editView(c *echo.Context, movie *models.Movie, form dto.AdminMovieForm) MovieFormView {
	return MovieFormView{
		Title:      "Edit movie",
		AdminEmail: adminEmail(c),
		CSRFToken:  csrfToken(c),
		Action:     moviesPath + "/" + strconv.FormatInt(movie.ID, 10),
		Submit:     "Save changes",
		Slug:       movie.Slug,
		IsEdit:     true,
		Form:       form,
	}
}

func (h *MovieHandler) renderForm(c *echo.Context, status int, view MovieFormView) error {
	genres, err := h.service.Genres(c.Request().Context())
	if err != nil {
		return utils.ServiceError(err)
	}
	if view.Errors == nil {
		view.Errors = map[string]string{}
	}
	view.Genres = genreOptions(genres, view.Form.GenreIDs)
	view.Cast = castRows(view.Form)
	view.AgeRatings = ageRatings
	view.Statuses = movieStatuses
	return c.Render(status, "admin/movies/form", view)
}

func (h *MovieHandler) renderFormError(c *echo.Context, view MovieFormView, err error) error {
	switch {
	case errors.Is(err, apperrors.ErrMovieGenreInvalid):
		view.Errors = map[string]string{"genre_ids": "One or more genres do not exist"}
	case errors.Is(err, apperrors.ErrMovieCastInvalid):
		view.Errors = map[string]string{"cast": "Each cast member needs an actor name"}
	case errors.Is(err, apperrors.ErrMovieModified):
		view.Error = "Someone else changed this movie while you were editing. Reload to see the latest data, or save again to overwrite it."
		return h.renderForm(c, http.StatusConflict, view)
	case errors.Is(err, apperrors.ErrMovieTokenInvalid):
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

func parsePage(raw string) int {
	page, err := strconv.Atoi(raw)
	if err != nil || page < 1 {
		return 1
	}
	return page
}

func toMovieRows(movies []models.Movie) []MovieRow {
	rows := make([]MovieRow, 0, len(movies))
	for _, movie := range movies {
		names := make([]string, 0, len(movie.Genres))
		for _, genre := range movie.Genres {
			names = append(names, genre.Name)
		}
		rows = append(rows, MovieRow{
			ID:          movie.ID,
			Title:       movie.Title,
			Slug:        movie.Slug,
			AgeRating:   string(movie.AgeRating),
			ReleaseDate: movie.ReleaseDate.Format(dateLayout),
			Status:      string(movie.Status),
			Genres:      strings.Join(names, ", "),
		})
	}
	return rows
}

func formFromMovie(movie *models.Movie) dto.AdminMovieForm {
	form := dto.AdminMovieForm{
		Title:         movie.Title,
		OriginalTitle: derefString(movie.OriginalTitle),
		Description:   derefString(movie.Description),
		DurationMin:   movie.DurationMin,
		AgeRating:     string(movie.AgeRating),
		Language:      derefString(movie.Language),
		Director:      derefString(movie.Director),
		PosterURL:     derefString(movie.PosterURL),
		BackdropURL:   derefString(movie.BackdropURL),
		TrailerURL:    derefString(movie.TrailerURL),
		ReleaseDate:   movie.ReleaseDate.Format(dateLayout),
		Status:        string(movie.Status),
		UpdatedAt:     updatedAtToken(movie.UpdatedAt),
	}
	for _, member := range movie.CastMembers {
		form.CastNames = append(form.CastNames, member.Name)
		form.CastRoles = append(form.CastRoles, member.Role)
	}
	for _, genre := range movie.Genres {
		form.GenreIDs = append(form.GenreIDs, genre.ID)
	}
	return form
}

func castRows(form dto.AdminMovieForm) []CastRow {
	n := max(len(form.CastNames), len(form.CastRoles))
	if n == 0 {
		return []CastRow{{}}
	}
	rows := make([]CastRow, 0, n)
	for i := 0; i < n; i++ {
		rows = append(rows, CastRow{Name: at(form.CastNames, i), Role: at(form.CastRoles, i)})
	}
	return rows
}

func genreOptions(genres []models.Genre, selected []int64) []GenreOption {
	options := make([]GenreOption, 0, len(genres))
	for _, genre := range genres {
		options = append(options, GenreOption{ID: genre.ID, Name: genre.Name, Checked: slices.Contains(selected, genre.ID)})
	}
	return options
}

func fieldErrors(err error) (map[string]string, string, bool) {
	var validationErr utils.ValidationError
	if errors.As(err, &validationErr) {
		fields := make(map[string]string, len(validationErr.Errors))
		for _, fieldErr := range validationErr.Errors {
			key := fieldKey(fieldErr.Field)
			if _, exists := fields[key]; !exists {
				fields[key] = fieldErr.Message
			}
		}
		return fields, "", true
	}
	var apiErr utils.APIErrorResponse
	if errors.As(err, &apiErr) && apiErr.ErrorCode == http.StatusBadRequest {
		return nil, "Please check the form", true
	}
	return nil, "", false
}

func fieldKey(field string) string {
	if i := strings.IndexByte(field, '['); i >= 0 {
		field = field[:i]
	}
	if strings.HasPrefix(field, "cast_") {
		return "cast"
	}
	return field
}

func at(values []string, i int) string {
	if i < len(values) {
		return values[i]
	}
	return ""
}

func updatedAtToken(t time.Time) string {
	return t.UTC().Format(time.RFC3339Nano)
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
