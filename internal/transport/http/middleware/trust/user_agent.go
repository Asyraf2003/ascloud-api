package trust

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

func UserAgentScore() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == http.MethodOptions {
				return next(c)
			}
			ua := strings.TrimSpace(c.Request().Header.Get("User-Agent"))
			if ua == "" {
				Add(c, -20, "ua_missing")
			} else {
				Add(c, 5, "ua_present")
			}
			return next(c)
		}
	}
}
