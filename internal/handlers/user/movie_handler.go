package user

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/services"
	"cinema-booking/internal/utils"
)

type MovieHandler struct {
	movies services.UserMovieService
}

func NewMovieHandler(movies services.UserMovieService) *MovieHandler {
	return &MovieHandler{movies: movies}
}

// @Summary List movies
// @Description Returns now-showing and coming-soon movies; `ended` movies are never listed. Without `status` both groups are returned. Title search combines substring matching with trigram word similarity, so a query of a few words with a typo or without Vietnamese diacritics still matches and the closest titles come first.
// @Tags movies
// @Produce json
// @Param status query string false "Filter by status" Enums(now_showing, coming_soon)
// @Param q query string false "Search by title (1–100 characters)"
// @Param page query int false "Page number, default 1" minimum(1) maximum(100000)
// @Param pageSize query int false "Items per page, default 20" minimum(1) maximum(50)
// @Success 200 {object} dto.MovieListResponse
// @Failure 400 {object} utils.ValidationError
// @Failure 500 {object} utils.APIErrorResponse
// @Router /movies [get]
func (h *MovieHandler) List(c *echo.Context) error {
	var query dto.MovieListQuery
	if err := utils.BindAndValidate(c, &query); err != nil {
		return err
	}

	page, err := h.movies.List(c.Request().Context(), query)
	if err != nil {
		return utils.ServiceError(err)
	}

	return c.JSON(http.StatusOK, dto.NewMovieListResponse(page.Movies, page.Page, page.PageSize, page.Total))
}

// @Summary Get a movie
// @Description Returns one movie with its genres, cast members and trailer. Ended movies are still returned; deleted or unknown ids are 404.
// @Tags movies
// @Produce json
// @Param id path int true "Movie ID"
// @Success 200 {object} dto.MovieDetail
// @Failure 404 {object} utils.APIErrorResponse "movie not found"
// @Failure 500 {object} utils.APIErrorResponse
// @Router /movies/{id} [get]
func (h *MovieHandler) Detail(c *echo.Context) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return err
	}

	movie, err := h.movies.Get(c.Request().Context(), id)
	if err != nil {
		return utils.ServiceError(err)
	}

	return c.JSON(http.StatusOK, dto.NewMovieDetail(movie))
}
