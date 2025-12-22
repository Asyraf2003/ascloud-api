package trust

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/shared/apperr"
)

func RequireOrigin(allowed []string) echo.MiddlewareFunc {
	allow := make(map[string]struct{}, len(allowed))
	for _, o := range allowed {
		allow[strings.TrimSpace(o)] = struct{}{}
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == http.MethodOptions {
				return next(c)
			}
			origin := strings.TrimSpace(c.Request().Header.Get("Origin"))
			if origin == "" {
				return apperr.New("FORBIDDEN", http.StatusForbidden, "Tidak punya akses.").
					WithField("reason", "origin_missing")
			}
			if _, ok := allow[origin]; !ok {
				return apperr.New("FORBIDDEN", http.StatusForbidden, "Tidak punya akses.").
					WithField("reason", "origin_not_allowed").
					WithField("origin", origin)
			}
			Add(c, 15, "origin_ok")
			return next(c)
		}
	}
}
