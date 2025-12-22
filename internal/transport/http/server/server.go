package server

import (
	"database/sql"
	"log/slog"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"example.com/your-api/internal/transport/http/middleware"
	"example.com/your-api/internal/transport/http/presenter"
)

func New(log *slog.Logger, db *sql.DB) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	e.Use(middleware.RequestID())
	e.Use(echomw.CORSWithConfig(corsConfig()))
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

func corsConfig() echomw.CORSConfig {
	origins := getenvList("AUTH_ALLOWED_ORIGINS", "http://localhost:8080")
	return echomw.CORSConfig{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "X-CSRF-Token", echo.HeaderXRequestID},
		ExposeHeaders:    []string{echo.HeaderXRequestID},
		AllowCredentials: true,
	}
}

func getenvList(k, def string) []string {
	raw := strings.TrimSpace(os.Getenv(k))
	if raw == "" {
		raw = def
	}
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
