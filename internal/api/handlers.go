package api

import (
	"context"
	"errors"
	"fmt"
	"kuchak/internal/config"
	"kuchak/internal/entity"
	"kuchak/pkg/auth"
	"kuchak/pkg/utils"
	"log"
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
	var loginRequest LoginRequest
	if err := c.Bind(&loginRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "invalid request body",
			},
		)
	}

	if err := c.Validate(loginRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	dbUser, err := w.App.AccountPostgres.GetUserByEmail(c.Request().Context(), loginRequest.Email)

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

	if !dbUser.IsEmailVerified {
		return c.JSON(
			http.StatusUnauthorized,
			echo.Map{
				"message": "email not verified",
			},
		)
	}

	if err := auth.PasswordVerify(dbUser.Password, loginRequest.Password); err != nil {
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
		AuthTokenResponse{
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
	var registerRequest RegisterRequest
	if err := c.Bind(&registerRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "invalid request body",
			},
		)
	}

	if err := c.Validate(registerRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	if strings.Compare(registerRequest.Password, registerRequest.PasswordRepeat) != 0 {
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "passwords is not match",
			},
		)
	}

	hashedPassword, err := auth.PasswordHash(registerRequest.Password)
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
		Email:    registerRequest.Email,
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

func (w *WebApp) requestVerifyEmail(c echo.Context) error {
	var emailRequest EmailRequest
	if err := c.Bind(&emailRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "invalid request body",
			},
		)
	}

	if err := c.Validate(emailRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	resp, err := w.App.AccountRedis.GetByVerifyEmail(c.Request().Context(), emailRequest.Email)
	if err != nil && rueidis.IsRedisNil(err) {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "email verification failed",
			},
		)
	}

	if err == nil && resp != "" {
		return c.JSON(
			http.StatusConflict,
			echo.Map{
				"message": "email verification already sent",
			},
		)
	}

	_, err = w.App.AccountPostgres.GetUserByEmail(c.Request().Context(), emailRequest.Email)

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

	token, err := auth.GenerateRandomToken(32)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to generate verify token",
		})
	}

	verifyEmailURL := fmt.Sprintf("%s/auth/verifyEmail/%s", w.appURL, token)

	if err := w.App.EmailSender.SendVerificationEmail(emailRequest.Email, verifyEmailURL); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to send email verification url",
		})
	}

	if err := w.App.AccountRedis.SetEmailVerify(c.Request().Context(), emailRequest.Email, token); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "email verification failed",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		echo.Map{
			"message": "verification email successfully sent",
		},
	)
}

func (w *WebApp) verifyEmail(c echo.Context) error {
	token := c.Param("token")

	email, err := w.App.AccountRedis.GetByVerifyToken(c.Request().Context(), token)
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
			http.StatusInternalServerError,
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
				"message": "failed to update email verify",
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

	var updateEmailRequest EmailRequest
	if err := c.Bind(&updateEmailRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "invalid request body",
			},
		)
	}

	if err := c.Validate(updateEmailRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	updatedUser := entity.User{
		ID:              user.UserID,
		Email:           updateEmailRequest.Email,
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

	var passwordUpdateRequest PasswordUpdateRequest
	if err := c.Bind(&passwordUpdateRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "invalid request body",
			},
		)
	}

	if err := c.Validate(passwordUpdateRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	if strings.Compare(passwordUpdateRequest.Password, passwordUpdateRequest.PasswordRepeat) != 0 {
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

	if err := auth.PasswordVerify(dbUser.Password, passwordUpdateRequest.OldPassword); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "incorrect password",
			},
		)
	}

	newPasswordHash, err := auth.PasswordHash(passwordUpdateRequest.Password)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to update password",
			},
		)
	}

	dbUser.Password = newPasswordHash

	if err := w.App.AccountPostgres.UpdateUserPassword(c.Request().Context(), dbUser); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusInternalServerError,
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
	var emailRequest EmailRequest
	if err := c.Bind(&emailRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "invalid request body",
			},
		)
	}

	if err := c.Validate(emailRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	resp, err := w.App.AccountRedis.GetByResetPasswordEmail(c.Request().Context(), emailRequest.Email)
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

	_, err = w.App.AccountPostgres.GetUserByEmail(c.Request().Context(), emailRequest.Email)

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

	token, err := auth.GenerateRandomToken(32)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to generate reset token",
		})
	}

	resetPasswordURL := fmt.Sprintf("%s/auth/resetPassword/%s", w.appURL, token)

	if err := w.App.EmailSender.SendResetPasswordEmail(emailRequest.Email, resetPasswordURL); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"message": "failed to send reset password url",
		})
	}

	if err := w.App.AccountRedis.SetResetPassword(c.Request().Context(), emailRequest.Email, token); err != nil {
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
	var passwordResetRequest PasswordResetRequest
	if err := c.Bind(&passwordResetRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "invalid request body",
			},
		)
	}

	if err := c.Validate(passwordResetRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	if strings.Compare(passwordResetRequest.Password, passwordResetRequest.PasswordRepeat) != 0 {
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "passwords is not match",
			},
		)
	}

	userEmail, err := w.App.AccountRedis.GetByResetPasswordToken(c.Request().Context(), passwordResetRequest.Token)
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
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to get user",
			},
		)
	}

	hashedPassword, err := auth.PasswordHash(passwordResetRequest.Password)
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

func (w *WebApp) createURL(c echo.Context) error {
	var createURLRequest URLRequest
	if err := c.Bind(&createURLRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "invalid request body",
			},
		)
	}

	if err := c.Validate(createURLRequest); err != nil {
		fmt.Printf("error: %v\n", err)
		return err
	}

	user := c.Get("user").(*auth.Claims)

	var newURL entity.URL
	var err error

	for {
		shortURL := utils.GenerateRandomString()
		newURL = entity.URL{
			ShortURL:    shortURL,
			OriginalURL: createURLRequest.OriginalURL,
			UserID:      user.UserID,
		}
		log.Printf("new url generated: %s, %v\n", shortURL, newURL)

		if err = w.App.URLPostgres.CreateURL(c.Request().Context(), newURL); err != nil {
			fmt.Printf("error: %v", err)
			if isUniqueViolation(err) {
				log.Printf("duplicate short URL, generating a new one\n")
				continue
			}

			return c.JSON(
				http.StatusInternalServerError,
				echo.Map{
					"message": "failed to create short url",
				},
			)
		}

		break
	}

	return c.JSON(
		http.StatusOK,
		newURL,
	)
}

func (w *WebApp) deleteURL(c echo.Context) error {
	shortURL := c.Param("shortURL")

	dbURL, err := w.App.URLPostgres.GetURLByShortURL(c.Request().Context(), shortURL)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Printf("error: %v\n", err)
			return c.JSON(
				http.StatusNotFound,
				echo.Map{
					"message": "url not found",
				},
			)
		}
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to get url",
			},
		)
	}

	user := c.Get("user").(*auth.Claims)

	if dbURL.UserID != user.UserID {
		return c.JSON(
			http.StatusForbidden,
			echo.Map{
				"message": "you not have access to delete this url",
			},
		)
	}

	if err = w.App.URLPostgres.DeleteURL(c.Request().Context(), dbURL); err != nil {
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to delete url",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		echo.Map{
			"message": "url deleted",
		},
	)
}

func (w *WebApp) getURL(c echo.Context) error {
	shortURL := c.Param("shortURL")

	dbURL, err := w.App.URLPostgres.GetURLByShortURL(c.Request().Context(), shortURL)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		if errors.Is(err, pgx.ErrNoRows) {
			fmt.Printf("error: %v\n", err)
			return c.JSON(
				http.StatusNotFound,
				echo.Map{
					"message": "url not found",
				},
			)
		}
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to get url",
			},
		)
	}

	user := c.Get("user").(*auth.Claims)

	if dbURL.UserID != user.UserID {
		return c.JSON(
			http.StatusForbidden,
			echo.Map{
				"message": "you not have access to get this url",
			},
		)
	}

	return c.JSON(
		http.StatusOK,
		dbURL,
	)
}

func (w *WebApp) getAllURLs(c echo.Context) error {
	user := c.Get("user").(*auth.Claims)

	urls, err := w.App.URLPostgres.GetURLsByUserID(c.Request().Context(), user.UserID)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to get urls",
			},
		)
	}

	if len(urls) == 0 {
		return c.JSON(
			http.StatusNotFound,
			echo.Map{
				"message": "urls is empty",
			},
		)
	}

	return c.JSON(http.StatusOK, urls)
}

func (w *WebApp) redirectURL(c echo.Context) error {
	shortURL := c.Param("shortURL")

	cacheURL, err := w.App.URLRedis.GetFromCacheByShortURL(c.Request().Context(), shortURL)
	if err == nil {
		fmt.Printf("redirected to: %s\n", cacheURL.OriginalURL)
		go func() {
			if err := w.App.URLPostgres.UpdateURLClickCount(context.Background(), shortURL); err != nil {
				fmt.Printf("DB error in click update: %v\n", err)
			}
			fmt.Printf("%s: click count updated\n", shortURL)
		}()
		return c.Redirect(http.StatusMovedPermanently, cacheURL.OriginalURL)
	}

	if !rueidis.IsRedisNil(err) {
		fmt.Printf("Redis error: %v\n", err)
	}

	dbURL, err := w.App.URLPostgres.GetURLByShortURL(c.Request().Context(), shortURL)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(
				http.StatusNotFound,
				echo.Map{
					"message": "url not found",
				},
			)
		}
		return c.JSON(
			http.StatusInternalServerError,
			echo.Map{
				"message": "failed to get url",
			},
		)
	}

	if err = w.App.URLRedis.SetURLToCache(c.Request().Context(), dbURL); err != nil {
		fmt.Printf("Redis set error: %v\n", err)
	}

	go func() {
		if err := w.App.URLPostgres.UpdateURLClickCount(context.Background(), shortURL); err != nil {
			fmt.Printf("DB error in click update: %v\n", err)
		}
		fmt.Printf("%s: click count updated\n", shortURL)
	}()

	fmt.Printf("redirected to: %s\n", dbURL.OriginalURL)
	return c.Redirect(http.StatusMovedPermanently, dbURL.OriginalURL)
}
