package handlers

import (
	"net/http"

	"github.com/labstack/echo/v5"

	"cinema-booking/internal/dto"
	"cinema-booking/internal/services"
	"cinema-booking/internal/utils"
)

type HealthHandler struct {
	healthService services.HealthService
}

func NewHealthHandler(healthService services.HealthService) *HealthHandler {
	return &HealthHandler{healthService: healthService}
}

// @Summary Health check
// @Description Returns 200 when the API can reach PostgreSQL, 503 otherwise.
// @Tags system
// @Produce json
// @Success 200 {object} dto.HealthResponse
// @Failure 503 {object} utils.APIErrorResponse
// @Router /health [get]
func (h *HealthHandler) Check(c *echo.Context) error {
	if err := h.healthService.Check(c.Request().Context()); err != nil {
		return utils.APIErrorFrom(http.StatusServiceUnavailable, "database unavailable", err)
	}

	return c.JSON(http.StatusOK, dto.HealthResponse{Status: "ok"})
}
