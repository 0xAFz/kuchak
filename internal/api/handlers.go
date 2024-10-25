package api

import (
	"errors"
	"fmt"
	"kuchak/internal/config"
	"kuchak/internal/entity"
	"kuchak/pkg/auth"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"github.com/redis/rueidis"
)

func (w *WebApp) healthz(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{
		"message": "ok",
	})
}

func (w *WebApp) login(c echo.Context) error {
	var loginUserData LoginUserData
	if err := c.Bind(&loginUserData); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "invalid request body",
			},
		)
	}

	if err := c.Validate(loginUserData); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	dbUser, err := w.App.AccountPostgres.GetUserByEmail(c.Request().Context(), loginUserData.Email)

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

	if err := auth.PasswordVerify(dbUser.Password, loginUserData.Password); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusUnauthorized,
			echo.Map{
				"message": "incorrect password",
			},
		)
	}

	accessToken, err := auth.GenerateToken(dbUser, config.AppConfig.AccessTokenSecret, auth.AccessTokenExp)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to generate access token",
			},
		)
	}

	refreshToken, err := auth.GenerateToken(dbUser, config.AppConfig.RefreshTokenSecret, auth.RefreshTokenExp)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to generate refresh token",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		TokenResponse{
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
		},
	)
}

func (w *WebApp) refreshToken(c echo.Context) error {
	refreshToken := c.Request().Header.Get("X-Refresh-Token")
	if refreshToken == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "refresh token required")
	}

	claims, err := auth.ValidateToken(refreshToken, config.AppConfig.RefreshTokenSecret)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
	}

	user := entity.User{
		ID:    claims.UserID,
		Email: claims.Email,
	}

	newAccessToken, err := auth.GenerateToken(user, config.AppConfig.AccessTokenSecret, auth.AccessTokenExp)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate access token")
	}

	return c.JSON(http.StatusOK, map[string]string{
		"access_token": newAccessToken,
	})
}

func (w *WebApp) register(c echo.Context) error {
	var registerUserData RegisterUserData
	if err := c.Bind(&registerUserData); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "invalid request body",
			},
		)
	}

	if err := c.Validate(registerUserData); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	if strings.Compare(registerUserData.Password, registerUserData.PasswordRepeat) != 0 {
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "passwords is not match",
			},
		)
	}

	hashedPassword, err := auth.PasswordHash(registerUserData.Password)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to create user",
			},
		)
	}

	newUser := entity.User{
		Email:    registerUserData.Email,
		Password: hashedPassword,
	}

	err = w.App.AccountPostgres.CreateUser(c.Request().Context(), newUser)

	if err != nil {
		fmt.Printf("error: %v\n", err)
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

func (w *WebApp) verifyEmail(c echo.Context) error {
	token := c.Param("token")

	email, err := w.App.AccountRedis.GetByEmailVerifyToken(c.Request().Context(), token)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		if errors.Is(err, rueidis.Nil) {
			return c.JSON(
				http.StatusBadRequest,
				echo.Map{
					"message": "token is not valid or expierd",
				},
			)
		}
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to verify token",
			},
		)
	}

	dbUser, err := w.App.AccountPostgres.GetUserByEmail(c.Request().Context(), email)

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
			http.StatusNotFound,
			echo.Map{
				"message": "failed to get user",
			},
		)
	}

	dbUser.IsEmailVerified = true

	if err := w.App.AccountPostgres.UpdateUserVerifyEmail(c.Request().Context(), dbUser); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to update email verify state",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		echo.Map{
			"message": "email verified",
		},
	)
}

func (w *WebApp) updateEmail(c echo.Context) error {
	user := c.Get("user").(*auth.Claims)

	var updateEmailData EmailData
	if err := c.Bind(&updateEmailData); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "invalid request body",
			},
		)
	}

	if err := c.Validate(updateEmailData); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	updatedUser := entity.User{
		ID:              user.UserID,
		Email:           updateEmailData.Email,
		IsEmailVerified: false,
	}

	if err := w.App.AccountPostgres.UpdateUserEmail(c.Request().Context(), updatedUser); err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to update email",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		echo.Map{
			"message":   "email updated",
			"new_email": updatedUser.Email,
		},
	)
}

func (w *WebApp) updatePassword(c echo.Context) error {
	user := c.Get("user").(*auth.Claims)

	var updatePasswordData UpdatePasswordData
	if err := c.Bind(&updatePasswordData); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "invalid request body",
			},
		)
	}

	if err := c.Validate(updatePasswordData); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	if strings.Compare(updatePasswordData.Password, updatePasswordData.PasswordRepeat) != 0 {
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "passwords is not match",
			},
		)
	}

	dbUser, err := w.App.AccountPostgres.GetUserByID(c.Request().Context(), user.UserID)

	if err != nil {
		fmt.Printf("error: %v\n", err)
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Printf("error: %v\n", err)
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

	if err := auth.PasswordVerify(dbUser.Password, updatePasswordData.OldPassword); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusUnauthorized,
			echo.Map{
				"message": "incorrect password",
			},
		)
	}

	newPasswordHash, err := auth.PasswordHash(updatePasswordData.Password)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusUnauthorized,
			echo.Map{
				"message": "failed to update password",
			},
		)
	}

	dbUser.Password = newPasswordHash

	if err := w.App.AccountPostgres.UpdateUserPassword(c.Request().Context(), dbUser); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusUnauthorized,
			echo.Map{
				"message": "failed to update password",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		echo.Map{
			"message": "password updated",
		},
	)

}

func (w *WebApp) requestResetPassword(c echo.Context) error {
	var emailData EmailData
	if err := c.Bind(&emailData); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "invalid request body",
			},
		)
	}

	if err := c.Validate(emailData); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	resp, err := w.App.AccountRedis.GetByResetPasswordEmail(c.Request().Context(), emailData.Email)
	if err != nil && rueidis.IsRedisNil(err) {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "reset password failed",
			},
		)
	}

	if err == nil && resp != "" {
		return c.JSON(
			http.StatusConflict,
			echo.Map{
				"message": "reset password already sent",
			},
		)
	}

	_, err = w.App.AccountPostgres.GetUserByEmail(c.Request().Context(), emailData.Email)

	if err != nil {
		fmt.Printf("error: %v\n", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(
				http.StatusBadRequest,
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

	// sendEmail(...)

	token, err := auth.GenerateRandomToken(32)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to generate reset token",
		})
	}

	if err := w.App.AccountRedis.SetResetPassword(c.Request().Context(), emailData.Email, token); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed reset password",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		echo.Map{
			"message": "reset password email successfully sent",
		},
	)
}

func (w *WebApp) resetPassword(c echo.Context) error {
	var resetPasswordData ResetPasswordData
	if err := c.Bind(&resetPasswordData); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "invalid request body",
			},
		)
	}

	if err := c.Validate(resetPasswordData); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	if strings.Compare(resetPasswordData.Password, resetPasswordData.PasswordRepeat) != 0 {
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "passwords is not match",
			},
		)
	}

	userEmail, err := w.App.AccountRedis.GetByResetPasswordToken(c.Request().Context(), resetPasswordData.Token)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		if rueidis.IsRedisNil(err) {
			return c.JSON(
				http.StatusBadRequest,
				echo.Map{
					"message": "token is not valid or expired",
				},
			)
		}
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to reset password",
			},
		)
	}

	dbUser, err := w.App.AccountPostgres.GetUserByEmail(c.Request().Context(), userEmail)
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
			http.StatusNotFound,
			echo.Map{
				"message": "failed to get user",
			},
		)
	}

	hashedPassword, err := auth.PasswordHash(resetPasswordData.Password)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to reset password",
			},
		)
	}

	dbUser.Password = hashedPassword

	if err := w.App.AccountPostgres.UpdateUserPassword(c.Request().Context(), dbUser); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to reset password",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		echo.Map{
			"message": "password changed successfully",
		},
	)
}
