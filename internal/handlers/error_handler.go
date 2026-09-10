package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	"cinema-booking/internal/middleware"
	"cinema-booking/internal/utils"
)

type adminErrorView struct {
	Title string
}

func HTTPErrorHandler(c *echo.Context, err error) {
	if response, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil && response.Committed {
		return
	}

	statusCode := statusCodeOf(err)
	if middleware.IsAdminPath(c.Request().URL.Path) {
		renderAdminError(c, statusCode)
		return
	}

	var marshaler json.Marshaler
	if errors.As(err, &marshaler) {
		writeJSON(c, statusCode, marshaler)
		return
	}

	writeJSON(c, statusCode, utils.APIError(statusCode, errorMessage(err, statusCode)))
}

func statusCodeOf(err error) int {
	var statusCoder echo.HTTPStatusCoder
	if errors.As(err, &statusCoder) && statusCoder.StatusCode() != 0 {
		return statusCoder.StatusCode()
	}
	return http.StatusInternalServerError
}

func errorMessage(err error, statusCode int) string {
	if statusCode >= http.StatusInternalServerError {
		return http.StatusText(statusCode)
	}
	var httpError *echo.HTTPError
	if errors.As(err, &httpError) && httpError.Message != "" {
		return httpError.Message
	}
	var apiError utils.APIErrorResponse
	if errors.As(err, &apiError) && apiError.ErrorMessage != "" {
		return apiError.ErrorMessage
	}
	return http.StatusText(statusCode)
}

func renderAdminError(c *echo.Context, statusCode int) {
	view := adminErrorView{Title: "Something went wrong"}
	if err := c.Render(statusCode, "admin/error", view); err != nil {
		c.Logger().Error("failed to render admin error page", "error", err)
		writeJSON(c, statusCode, utils.APIError(statusCode, http.StatusText(statusCode)))
	}
}

func writeJSON(c *echo.Context, statusCode int, body any) {
	if err := c.JSON(statusCode, body); err != nil {
		c.Logger().Error("failed to write error response", "error", err)
	}
}
