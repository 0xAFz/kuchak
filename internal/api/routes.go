package api

import (
	"kuchak/internal/config"
	"kuchak/pkg/auth"
	"kuchak/pkg/validate"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
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

	config := echojwt.Config{
		NewClaimsFunc: func(c echo.Context) jwt.Claims {
			return new(auth.JwtClaims)
		},
		SigningKey: []byte(config.AppConfig.SecretKey),
	}

	r := w.e.Group("/restricted")
	r.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "OK\n")
	})

	r.Use(echojwt.WithConfig(config))

	a := w.e.Group("/auth")
	a.POST("/login", w.login)
	a.POST("/signup", w.signup)

	w.e.GET("/healthz", w.healthz)
}
