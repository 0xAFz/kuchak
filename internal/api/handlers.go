package api

import (
	"kuchak/pkg/auth"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (w *WebApp) healthz(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{
		"status": "ok",
	})
}

func (w *WebApp) login(c echo.Context) error {
	var userData InputUser
	if err := c.Bind(&userData); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"error": "Bad request",
			},
		)
	}

	t, err := auth.GenerateJwtToken(userData.Email)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"error": "Failed to generate jwt token",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		echo.Map{
			"token": t,
		},
	)
}

// func (w *WebApp) signup(c echo.Context) error {

// }
