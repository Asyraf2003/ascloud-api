package trust

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/transport/http/middleware"
)

func ScoreFromAAL() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == http.MethodOptions {
				return next(c)
			}
			got, _ := middleware.AALLevel(c)
			switch strings.ToLower(strings.TrimSpace(got)) {
			case "aal3":
				Add(c, 20, "aal3")
			case "aal2":
				Add(c, 10, "aal2")
			case "aal1":
				Add(c, -10, "aal1")
			default:
				Add(c, -20, "aal_unknown")
			}
			return next(c)
		}
	}
}
