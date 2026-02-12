package server

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"example.com/your-api/internal/transport/http/middleware"
	"example.com/your-api/internal/transport/http/presenter"
)

func New(log *slog.Logger, db any, allowedOrigins []string) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.RequestID())
	e.Use(echomw.CORSWithConfig(corsConfig(allowedOrigins)))
	e.Use(middleware.AccessLog(log))
	e.Use(echomw.Recover())

	e.HTTPErrorHandler = presenter.HTTPErrorHandler

	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("logger", log)
			c.Set("db", db)
			return next(c)
		}
	})

	return e
}

func corsConfig(origins []string) echomw.CORSConfig {
	return echomw.CORSConfig{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-CSRF-Token", echo.HeaderXRequestID},
		ExposeHeaders:    []string{echo.HeaderXRequestID},
		AllowCredentials: true,
	}
}
