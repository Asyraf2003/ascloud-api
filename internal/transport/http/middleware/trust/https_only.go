package trust

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/shared/apperr"
)

func RequireHTTPS() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == http.MethodOptions {
				return next(c)
			}
			if isHTTPS(c) {
				Add(c, 10, "https")
				return next(c)
			}
			if isProd() {
				return apperr.New("FORBIDDEN", http.StatusForbidden, "Tidak punya akses.").
					WithField("reason", "https_required")
			}
			Add(c, -10, "http_dev")
			return next(c)
		}
	}
}

func isProd() bool {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if env == "" {
		env = "dev"
	}
	return env != "dev"
}

func isHTTPS(c echo.Context) bool {
	if c.Request().TLS != nil {
		return true
	}
	if strings.EqualFold(c.Request().Header.Get("X-Forwarded-Proto"), "https") {
		return true
	}
	if strings.EqualFold(c.Request().Header.Get("X-Forwarded-SSL"), "on") {
		return true
	}
	var v struct {
		Scheme string `json:"scheme"`
	}
	if err := json.Unmarshal([]byte(c.Request().Header.Get("CF-Visitor")), &v); err == nil && strings.EqualFold(v.Scheme, "https") {
		return true
	}
	return false
}
