package server

import (
	"database/sql"
	"log/slog"

	"github.com/labstack/echo/v4"
	echomw "github.com/labstack/echo/v4/middleware"

	"example.com/your-api/internal/transport/http/middleware"
	"example.com/your-api/internal/transport/http/presenter"
)

// New membuat instance Echo + middleware global.
// db boleh nil (fase awal).
func New(log *slog.Logger, db *sql.DB) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.HidePort = true

	// Global middleware (urutan penting)
	e.Use(middleware.RequestID())
	e.Use(middleware.AccessLog(log))
	e.Use(echomw.Recover())

	// Semua error response tersanitasi di sini
	e.HTTPErrorHandler = presenter.HTTPErrorHandler

	// Attach dependencies (fase awal). Nanti kalau udah gede, pindah ke DI/wire.
	e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			c.Set("logger", log)
			c.Set("db", db)
			return next(c)
		}
	})

	return e
}
