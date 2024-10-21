package api

import (
	"kuchak/pkg/auth"
	"net/http"

	"github.com/labstack/echo/v4"
)

func (w *WebApp) healthz(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{
		"message": "ok",
	})
}

func (w *WebApp) login(c echo.Context) error {
	var userData InputUser
	if err := c.Bind(&userData); err != nil {
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "invalid request body",
			},
		)
	}

	if err := c.Validate(userData); err != nil {
		return err
	}

	t, err := auth.GenerateJwtToken(userData.Email)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to generate token",
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
