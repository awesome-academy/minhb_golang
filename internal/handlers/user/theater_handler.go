package user

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/services"
	"cinema-booking/internal/utils"
)

type TheaterHandler struct {
	theaters services.UserTheaterService
}

func NewTheaterHandler(theaters services.UserTheaterService) *TheaterHandler {
	return &TheaterHandler{theaters: theaters}
}

// @Summary List theaters
// @Description Returns every active theater ordered by city then name. Inactive or deleted theaters are never listed. No paging.
// @Tags theaters
// @Produce json
// @Param city query string false "Filter by city (case-insensitive exact match)" maxLength(100)
// @Success 200 {object} dto.TheaterListResponse
// @Failure 400 {object} utils.ValidationError
// @Failure 500 {object} utils.APIErrorResponse
// @Router /theaters [get]
func (h *TheaterHandler) List(c *echo.Context) error {
	var query dto.TheaterListQuery
	if err := utils.BindAndValidate(c, &query); err != nil {
		return err
	}

	theaters, err := h.theaters.List(c.Request().Context(), query.City)
	if err != nil {
		return utils.ServiceError(err)
	}

	return c.JSON(http.StatusOK, dto.NewTheaterListResponse(theaters))
}

// @Summary Get a theater
// @Description Returns one active theater. Inactive, deleted or unknown ids are 404.
// @Tags theaters
// @Produce json
// @Param id path int true "Theater ID"
// @Success 200 {object} dto.TheaterResponse
// @Failure 404 {object} utils.APIErrorResponse "theater not found or inactive"
// @Failure 500 {object} utils.APIErrorResponse
// @Router /theaters/{id} [get]
func (h *TheaterHandler) Detail(c *echo.Context) error {
	id, err := utils.ParamID(c, "id")
	if err != nil {
		return err
	}

	theater, err := h.theaters.Get(c.Request().Context(), id)
	if err != nil {
		return utils.ServiceError(err)
	}

	return c.JSON(http.StatusOK, dto.NewTheaterResponse(theater))
}
