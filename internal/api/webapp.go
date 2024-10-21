package api

import (
	"context"
	"kuchak/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type WebApp struct {
	addr string
	App  *service.App
	e    *echo.Echo
}

func NewWebApp(
	addr string,
	app *service.App,
) *WebApp {
	e := echo.New()
	wa := &WebApp{
		App:  app,
		e:    e,
		addr: addr,
	}
	wa.routes()
	return wa
}

func (w *WebApp) Start() error {
	w.e.Use(middleware.Recover())
	return w.e.Start(w.addr)
}

func (w *WebApp) Shutdown(ctx context.Context) error {
	return w.e.Shutdown(ctx)
}
