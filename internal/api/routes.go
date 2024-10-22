package api

import (
	"kuchak/internal/config"
	"kuchak/pkg/auth"
	"kuchak/pkg/validate"
	"net/http"

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
	a.POST("/login", w.login)
	a.POST("/register", w.register)
	a.POST("/refresh", w.refreshToken)
	a.POST("/updateEmail", w.updateEmail, w.authMiddleware)
	a.POST("/updatePassword", w.updatePassword, w.authMiddleware)
	a.GET("/verify/:token", w.verifyEmail)

	w.e.GET("/healthz", w.healthz)
}

func (w *WebApp) authMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
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
