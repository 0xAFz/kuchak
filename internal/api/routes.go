package api

import (
	"kuchak/internal/config"
	"kuchak/pkg/auth"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func (w *WebApp) routes() {
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

	w.e.GET("/healthz", w.healthz)
	w.e.POST("/login", w.login)
}
