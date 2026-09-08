package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/labstack/echo/v5"

	"cinema-booking/internal/utils"
)

func HTTPErrorHandler(c *echo.Context, err error) {
	if response, unwrapErr := echo.UnwrapResponse(c.Response()); unwrapErr == nil && response.Committed {
		return
	}

	statusCode := http.StatusInternalServerError
	var statusCoder echo.HTTPStatusCoder
	if errors.As(err, &statusCoder) && statusCoder.StatusCode() != 0 {
		statusCode = statusCoder.StatusCode()
	}

	var marshaler json.Marshaler
	if errors.As(err, &marshaler) {
		writeJSON(c, statusCode, marshaler)
		return
	}

	message := http.StatusText(statusCode)
	var httpError *echo.HTTPError
	if errors.As(err, &httpError) && httpError.Message != "" && statusCode < http.StatusInternalServerError {
		message = httpError.Message
	}

	writeJSON(c, statusCode, utils.APIError(statusCode, message))
}

func writeJSON(c *echo.Context, statusCode int, body any) {
	if err := c.JSON(statusCode, body); err != nil {
		c.Logger().Error("failed to write error response", "error", err)
	}
}
