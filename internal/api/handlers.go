package api

import (
	"context"
	"errors"
	"fmt"
	"kuchak/internal/config"
	"kuchak/internal/entity"
	"kuchak/pkg/auth"
	"kuchak/pkg/utils"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/labstack/echo/v4"
	"github.com/redis/rueidis"
	"github.com/rs/zerolog/log"
)

func (w *WebApp) healthz(c echo.Context) error {
	return c.String(http.StatusOK, "OK\n")
}

func (w *WebApp) login(c echo.Context) error {
	var loginRequest LoginRequest
	if err := c.Bind(&loginRequest); err != nil {
		log.Err(err).Msg("failed to bind request body")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: "invalid request body",
			Success: false,
		})
	}

	if err := c.Validate(loginRequest); err != nil {
		log.Err(err).Msg("failed to validate payload")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: fmt.Sprintf("failed to validate payload: %s", err.Error()),
			Success: false,
		})
	}

	dbUser, err := w.App.AccountPostgres.GetUserByEmail(c.Request().Context(), loginRequest.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrMessage{
				Message: "user not found",
				Success: false,
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to fetch user",
			Success: false,
		})
	}

	if !dbUser.IsEmailVerified {
		log.Error().Str("email", dbUser.Email).Msg("email not verified")
		return c.JSON(http.StatusUnauthorized, ErrMessage{
			Message: "email not verified",
			Success: false,
		})
	}

	if err := auth.PasswordVerify(dbUser.Password, loginRequest.Password); err != nil {
		return c.JSON(http.StatusUnauthorized, ErrMessage{
			Message: "incorrect password",
			Success: false,
		})
	}

	accessToken, err := auth.GenerateToken(dbUser, config.AppConfig.AccessTokenSecret, auth.AccessTokenExp)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to generate access token",
			Success: false,
		})
	}

	refreshToken, err := auth.GenerateToken(dbUser, config.AppConfig.RefreshTokenSecret, auth.RefreshTokenExp)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to generate refresh token",
			Success: false,
		})
	}

	return c.JSON(http.StatusOK, AuthTokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
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

	return c.JSON(http.StatusOK, echo.Map{
		"access_token": newAccessToken,
	})
}

func (w *WebApp) register(c echo.Context) error {
	var registerRequest RegisterRequest
	if err := c.Bind(&registerRequest); err != nil {
		log.Err(err).Msg("failed to bind request body")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: "invalid request body",
			Success: false,
		})
	}

	if err := c.Validate(registerRequest); err != nil {
		log.Err(err).Msg("failed to validate payload")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: fmt.Sprintf("failed to validate payload: %s", err.Error()),
			Success: false,
		})
	}

	if registerRequest.Password != registerRequest.PasswordRepeat {
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: "passwords is not match",
			Success: false,
		})
	}

	dbUser, err := w.App.AccountPostgres.GetUserByEmail(c.Request().Context(), registerRequest.Email)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to create user",
			Success: false,
		})
	}

	if dbUser.Email != "" {
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: "user already exists",
			Success: false,
		})
	}

	hashedPassword, err := auth.PasswordHash(registerRequest.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to create user",
			Success: false,
		})
	}

	newUser := entity.User{
		Email:    registerRequest.Email,
		Password: hashedPassword,
	}

	err = w.App.AccountPostgres.CreateUser(c.Request().Context(), newUser)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to create user",
			Success: false,
		})
	}

	return c.JSON(http.StatusOK, ResponseOk{
		Message: "user created",
		Success: true,
		Data: echo.Map{
			"email": newUser.Email,
		},
	})
}

func (w *WebApp) requestVerifyEmail(c echo.Context) error {
	var emailRequest EmailRequest
	if err := c.Bind(&emailRequest); err != nil {
		log.Err(err).Msg("failed to bind request body")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: "invalid request body",
			Success: false,
		})
	}

	if err := c.Validate(emailRequest); err != nil {
		log.Err(err).Msg("failed to validate payload")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: fmt.Sprintf("failed to validate payload: %s", err.Error()),
			Success: false,
		})
	}

	resp, err := w.App.AccountRedis.GetByVerifyEmail(c.Request().Context(), emailRequest.Email)
	if err != nil && !errors.Is(err, rueidis.Nil) {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "email verification failed",
			Success: false,
		})
	}

	if resp != "" {
		return c.JSON(http.StatusConflict, ErrMessage{
			Message: "email verification already sent",
			Success: false,
		})
	}

	_, err = w.App.AccountPostgres.GetUserByEmail(c.Request().Context(), emailRequest.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusBadRequest, ErrMessage{
				Message: "user not found",
				Success: false,
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to fetch user",
		})
	}

	token, err := auth.GenerateRandomToken(32)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to generate verify token",
			Success: false,
		})
	}

	verifyEmailURL := fmt.Sprintf("%s/auth/verifyEmail/%s", w.appURL, token)

	if err := w.App.EmailSender.SendVerificationEmail(emailRequest.Email, verifyEmailURL); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to send email verification url",
			Success: false,
		})
	}

	if err := w.App.AccountRedis.SetEmailVerify(c.Request().Context(), emailRequest.Email, token); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "email verification failed",
		})
	}

	return c.JSON(http.StatusOK, ResponseOk{
		Message: "verification email successfully sent",
		Success: true,
	})
}

func (w *WebApp) verifyEmail(c echo.Context) error {
	token := c.Param("token")

	email, err := w.App.AccountRedis.GetByVerifyToken(c.Request().Context(), token)
	if err != nil {
		if errors.Is(err, rueidis.Nil) {
			return c.JSON(http.StatusBadRequest, ErrMessage{
				Message: "verify token is not valid or expierd",
				Success: false,
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to check verify token",
			Success: false,
		})
	}

	dbUser, err := w.App.AccountPostgres.GetUserByEmail(c.Request().Context(), email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrMessage{
				Message: "user not found",
				Success: false,
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to fetch user",
			Success: false,
		})
	}

	dbUser.IsEmailVerified = true

	if err := w.App.AccountPostgres.UpdateUserVerifyEmail(c.Request().Context(), dbUser); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to verify email",
		})
	}

	return c.JSON(http.StatusOK, ResponseOk{
		Message: "email verified successfully",
		Success: true,
		Data: echo.Map{
			"user": dbUser,
		},
	})
}

func (w *WebApp) updateEmail(c echo.Context) error {
	var updateEmailRequest EmailRequest
	if err := c.Bind(&updateEmailRequest); err != nil {
		log.Err(err).Msg("failed to bind request body")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: "invalid request body",
			Success: false,
		})
	}

	if err := c.Validate(updateEmailRequest); err != nil {
		log.Err(err).Msg("failed to validate payload")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: fmt.Sprintf("failed to validate payload: %s", err.Error()),
			Success: false,
		})
	}

	user := c.Get("user").(*auth.Claims)

	updatedUser := entity.User{
		ID:              user.UserID,
		Email:           updateEmailRequest.Email,
		IsEmailVerified: false,
	}

	if err := w.App.AccountPostgres.UpdateUserEmail(c.Request().Context(), updatedUser); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to update email",
			Success: false,
		})
	}

	return c.JSON(http.StatusOK, ResponseOk{
		Message: "email updated successfully",
		Success: true,
		Data: echo.Map{
			"user": updatedUser,
		},
	})
}

func (w *WebApp) updatePassword(c echo.Context) error {
	var passwordUpdateRequest PasswordUpdateRequest
	if err := c.Bind(&passwordUpdateRequest); err != nil {
		log.Err(err).Msg("failed to bind request body")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: "invalid request body",
			Success: false,
		})
	}

	if err := c.Validate(passwordUpdateRequest); err != nil {
		log.Err(err).Msg("failed to validate payload")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: fmt.Sprintf("failed to validate payload: %s", err.Error()),
			Success: false,
		})
	}

	if passwordUpdateRequest.Password != passwordUpdateRequest.PasswordRepeat {
		return c.JSON(
			http.StatusBadRequest,
			echo.Map{
				"message": "passwords is not match",
			},
		)
	}

	user := c.Get("user").(*auth.Claims)

	dbUser, err := w.App.AccountPostgres.GetUserByID(c.Request().Context(), user.UserID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrMessage{
				Message: "user not found",
				Success: false,
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to fetch user",
			Success: false,
		})
	}

	if err := auth.PasswordVerify(dbUser.Password, passwordUpdateRequest.OldPassword); err != nil {
		return c.JSON(http.StatusUnauthorized, ErrMessage{
			Message: "incorrect password",
			Success: false,
		})
	}

	newPasswordHash, err := auth.PasswordHash(passwordUpdateRequest.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to update password",
			Success: false,
		})
	}

	dbUser.Password = newPasswordHash

	if err := w.App.AccountPostgres.UpdateUserPassword(c.Request().Context(), dbUser); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to update password",
			Success: false,
		})
	}

	return c.JSON(http.StatusOK, ResponseOk{
		Message: "password updated successfully",
		Success: true,
	})
}

func (w *WebApp) requestResetPassword(c echo.Context) error {
	var emailRequest EmailRequest
	if err := c.Bind(&emailRequest); err != nil {
		log.Err(err).Msg("failed to bind request body")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: "invalid request body",
			Success: false,
		})
	}

	if err := c.Validate(emailRequest); err != nil {
		log.Err(err).Msg("failed to validate payload")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: fmt.Sprintf("failed to validate payload: %s", err.Error()),
			Success: false,
		})
	}

	resp, err := w.App.AccountRedis.GetByResetPasswordEmail(c.Request().Context(), emailRequest.Email)
	if err != nil && !errors.Is(err, rueidis.Nil) {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "request reset password failed",
			Success: false,
		})
	}

	if resp != "" {
		return c.JSON(http.StatusConflict, ErrMessage{
			Message: "reset password already sent",
			Success: false,
		})
	}

	_, err = w.App.AccountPostgres.GetUserByEmail(c.Request().Context(), emailRequest.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusBadRequest, ErrMessage{
				Message: "user not found",
				Success: false,
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to fetch user",
			Success: false,
		})
	}

	token, err := auth.GenerateRandomToken(32)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to generate reset password token",
			Success: false,
		})
	}

	resetPasswordURL := fmt.Sprintf("%s/auth/resetPassword/%s", w.appURL, token)

	if err := w.App.EmailSender.SendResetPasswordEmail(emailRequest.Email, resetPasswordURL); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to send reset password url",
			Success: false,
		})
	}

	if err := w.App.AccountRedis.SetResetPassword(c.Request().Context(), emailRequest.Email, token); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "request reset password failed",
			Success: false,
		})
	}

	return c.JSON(http.StatusOK, ResponseOk{
		Message: "reset password email successfully sent",
		Success: true,
	})
}

func (w *WebApp) resetPassword(c echo.Context) error {
	var passwordResetRequest PasswordResetRequest
	if err := c.Bind(&passwordResetRequest); err != nil {
		log.Err(err).Msg("failed to bind request body")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: "invalid request body",
			Success: false,
		})
	}

	if err := c.Validate(passwordResetRequest); err != nil {
		log.Err(err).Msg("failed to validate payload")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: fmt.Sprintf("failed to validate payload: %s", err.Error()),
			Success: false,
		})
	}

	if passwordResetRequest.Password != passwordResetRequest.PasswordRepeat {
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: "passwords is not match",
			Success: false,
		})
	}

	userEmail, err := w.App.AccountRedis.GetByResetPasswordToken(c.Request().Context(), passwordResetRequest.Token)
	if err != nil {
		if errors.Is(err, rueidis.Nil) {
			return c.JSON(http.StatusBadRequest, ErrMessage{
				Message: "reset password token is not valid or expired",
				Success: false,
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to reset password",
			Success: false,
		})
	}

	dbUser, err := w.App.AccountPostgres.GetUserByEmail(c.Request().Context(), userEmail)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrMessage{
				Message: "user not found",
				Success: false,
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to fetch user",
			Success: false,
		})
	}

	hashedPassword, err := auth.PasswordHash(passwordResetRequest.Password)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to reset password",
			Success: false,
		})
	}

	dbUser.Password = hashedPassword

	if err := w.App.AccountPostgres.UpdateUserPassword(c.Request().Context(), dbUser); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to reset password",
			Success: false,
		})
	}

	return c.JSON(http.StatusOK, ResponseOk{
		Message: "password changed successfully",
		Success: true,
	})
}

func isUniqueViolation(err error) bool {
	if pgErr, ok := err.(*pgconn.PgError); ok {
		return pgErr.Code == "23505"
	}
	return false
}

func (w *WebApp) createURL(c echo.Context) error {
	var createURLRequest URLRequest
	if err := c.Bind(&createURLRequest); err != nil {
		log.Err(err).Msg("failed to bind request body")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: "invalid request body",
			Success: false,
		})
	}

	if err := c.Validate(createURLRequest); err != nil {
		log.Err(err).Msg("failed to validate payload")
		return c.JSON(http.StatusBadRequest, ErrMessage{
			Message: fmt.Sprintf("failed to validate payload: %s", err.Error()),
			Success: false,
		})
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
		log.Info().Str("short_url", shortURL).Msg("new url generated")

		if err = w.App.URLPostgres.CreateURL(c.Request().Context(), newURL); err != nil {
			if isUniqueViolation(err) {
				log.Info().Str("short_url", shortURL).Msg("duplicate short url, generating a new one")
				continue
			}

			return c.JSON(http.StatusInternalServerError, ErrMessage{
				Message: "failed to create url",
				Success: false,
			})
		}

		break
	}

	return c.JSON(http.StatusOK, ResponseOk{
		Message: "url created successfully",
		Success: true,
		Data: echo.Map{
			"url": newURL,
		},
	})
}

func (w *WebApp) deleteURL(c echo.Context) error {
	shortURL := c.Param("shortURL")

	dbURL, err := w.App.URLPostgres.GetURLByShortURL(c.Request().Context(), shortURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrMessage{
				Message: "url not found",
				Success: false,
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to fetch url",
			Success: false,
		})
	}

	user := c.Get("user").(*auth.Claims)

	if dbURL.UserID != user.UserID {
		return c.JSON(http.StatusForbidden, ErrMessage{
			Message: "not have access to delete this url",
			Success: false,
		})
	}

	if err = w.App.URLPostgres.DeleteURL(c.Request().Context(), dbURL); err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to delete url",
			Success: false,
		})
	}

	return c.JSON(http.StatusOK, ResponseOk{
		Message: "url deleted successfully",
		Success: true,
	})
}

func (w *WebApp) getURL(c echo.Context) error {
	shortURL := c.Param("shortURL")

	dbURL, err := w.App.URLPostgres.GetURLByShortURL(c.Request().Context(), shortURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrMessage{
				Message: "url not found",
				Success: false,
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to fetch url",
			Success: false,
		})
	}

	user := c.Get("user").(*auth.Claims)

	if dbURL.UserID != user.UserID {
		return c.JSON(http.StatusForbidden, ErrMessage{
			Message: "not have access to fetch this url",
			Success: false,
		})
	}

	return c.JSON(http.StatusOK, ResponseOk{
		Success: true,
		Data: echo.Map{
			"url": dbURL,
		},
	})
}

func (w *WebApp) getAllURLs(c echo.Context) error {
	user := c.Get("user").(*auth.Claims)

	urls, err := w.App.URLPostgres.GetURLsByUserID(c.Request().Context(), user.UserID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to fetch urls",
			Success: false,
		})
	}

	if len(urls) == 0 {
		return c.JSON(http.StatusNotFound, ErrMessage{
			Message: "urls is empty",
			Success: false,
		})
	}

	return c.JSON(http.StatusOK, ResponseOk{
		Success: true,
		Data: echo.Map{
			"urls": urls,
		},
	})
}

func (w *WebApp) redirectURL(c echo.Context) error {
	shortURL := c.Param("shortURL")

	cacheURL, err := w.App.URLRedis.GetFromCacheByShortURL(c.Request().Context(), shortURL)
	if err == nil {
		log.Info().Str("short_url", shortURL).Msg("redirected from cache")

		go w.App.URLPostgres.UpdateURLClickCount(context.Background(), shortURL)

		return c.Redirect(http.StatusMovedPermanently, cacheURL.OriginalURL)
	}

	dbURL, err := w.App.URLPostgres.GetURLByShortURL(c.Request().Context(), shortURL)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return c.JSON(http.StatusNotFound, ErrMessage{
				Message: "url not found",
				Success: false,
			})
		}
		return c.JSON(http.StatusInternalServerError, ErrMessage{
			Message: "failed to fetch url",
			Success: false,
		})
	}

	w.App.URLRedis.SetURLToCache(c.Request().Context(), dbURL)

	go w.App.URLPostgres.UpdateURLClickCount(context.Background(), shortURL)

	log.Info().Str("short_url", shortURL).Msg("redirected from db")

	return c.Redirect(http.StatusMovedPermanently, dbURL.OriginalURL)
}
