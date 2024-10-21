package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (w *WebApp) healthz(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{
		"status": "ok",
	})
}
