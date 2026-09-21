package utils

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
)

func ParamID(c *echo.Context, name string) (int64, error) {
	id, err := strconv.ParseInt(c.Param(name), 10, 64)
	if err != nil || id <= 0 {
		return 0, APIError(http.StatusNotFound, "resource not found")
	}
	return id, nil
}
