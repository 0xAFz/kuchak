package api

import (
	"kuchak/internal/config"
	"kuchak/pkg/auth"
	"kuchak/pkg/validate"
	"net/http"
	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (w *WebApp) routes() {
	v := validator.New()
	v.RegisterValidation("password", validate.CustomPasswordValidator)

	w.e.Validator = &validate.CustomValidator{Validator: v}

	w.e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodOptions,
			http.MethodHead,
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
		},
		AllowHeaders: []string{
			"*",
			echo.HeaderAuthorization,
		},
		AllowCredentials: true,
	}))

	a := w.e.Group("/auth")
	a.Use(w.rateLimit(20, time.Hour*2))
	a.POST("/login", w.login)
	a.POST("/register", w.register)
	a.POST("/refresh", w.refreshToken)
	a.POST("/requestResetPassword", w.requestResetPassword)
	a.POST("/resetPassword", w.resetPassword)
	a.PATCH("/updateEmail", w.updateEmail, w.withAuth())
	a.PATCH("/updatePassword", w.updatePassword, w.withAuth())
	a.POST("/requestVerifyEmail", w.requestVerifyEmail)
	a.GET("/verifyEmail/:token", w.verifyEmail)

	u := w.e.Group("/urls")
	u.Use(w.rateLimit(100, time.Hour*2))
	u.Use(w.withAuth())
	u.GET("/get/:shortURL", w.getURL)
	u.GET("/getAll", w.getAllURLs)
	u.POST("/create", w.createURL)
	u.DELETE("/delete/:shortURL", w.deleteURL)

	w.e.GET("/healthz", w.healthz, w.rateLimit(2, time.Hour*1))
	w.e.GET("/:shortURL", w.redirectURL)
}

func (w *WebApp) withAuth() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return echo.NewHTTPError(http.StatusUnauthorized, "missing authorization header")
			}

			tokenStr := authHeader[len("Bearer "):]

			claims, err := auth.ValidateToken(tokenStr, config.AppConfig.AccessTokenSecret)
			if err != nil {
				return echo.NewHTTPError(http.StatusUnauthorized, "invalid token")
			}

			c.Set("user", claims)
			return next(c)
		}
	}
}

func (w *WebApp) rateLimit(limit int, window time.Duration) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			ip := c.RealIP()

			allowed, remaining, reset, err := w.App.RateLimit.IsAllowed(c.Request().Context(), ip, limit, window)
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, "rate limit check failed")
			}

			c.Response().Header().Set("X-RateLimit-Limit", strconv.Itoa(limit))
			c.Response().Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
			c.Response().Header().Set("X-RateLimit-Reset", strconv.FormatInt(reset.Unix(), 10))

			if !allowed {
				return echo.NewHTTPError(http.StatusTooManyRequests, "too many requests")
			}

			return next(c)
		}
	}
}
