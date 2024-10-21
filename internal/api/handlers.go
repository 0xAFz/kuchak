package api

import (
	"errors"
	"fmt"
	"kuchak/internal/entity"
	"kuchak/pkg/auth"
	"net/http"

	"github.com/jackc/pgx/v5"
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

	dbUser, err := w.App.AccountPostgres.GetUserByEmail(c.Request().Context(), userData.Email)

	if err != nil {
		fmt.Printf("error: %v\n", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(
				http.StatusNotFound,
				echo.Map{
					"message": "user not found",
				},
			)
		}
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to get user",
			},
		)
	}

	if err := auth.PasswordVerify(dbUser.Password, userData.Password); err != nil {
		return c.JSON(
			http.StatusUnauthorized,
			echo.Map{
				"message": "incorrect password",
			},
		)
	}

	t, err := auth.GenerateJwtToken(dbUser.Email)
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

func (w *WebApp) signup(c echo.Context) error {
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

	hashedPassword, err := auth.PasswordHash(userData.Password)
	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to create user",
			},
		)
	}

	newUser := entity.User{
		Email:    userData.Email,
		Password: hashedPassword,
	}

	err = w.App.AccountPostgres.CreateUser(c.Request().Context(), newUser)

	if err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to create user",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		echo.Map{
			"message": "user created",
		},
	)
}
