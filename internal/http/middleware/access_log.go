package middleware

import (
	"log/slog"
	"time"

	"github.com/labstack/echo/v4"
)

func AccessLog(log *slog.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			// IMPORTANT:
			// Kalau handler return error dan response belum committed,
			// Echo baru akan panggil HTTPErrorHandler setelah middleware chain balik.
			// Jadi kita panggil c.Error(err) di sini biar status final sudah ke-set,
			// lalu return nil supaya tidak diproses dua kali.
			if err != nil && !c.Response().Committed {
				c.Error(err)
				err = nil
			}

			req := c.Request()
			rid := c.Response().Header().Get(echo.HeaderXRequestID)

			log.Info("http",
				"request_id", rid,
				"method", req.Method,
				"path", req.URL.Path,
				"status", c.Response().Status,
				"latency_ms", time.Since(start).Milliseconds(),
				"remote_ip", c.RealIP(),
			)

			return err
		}
	}
}
