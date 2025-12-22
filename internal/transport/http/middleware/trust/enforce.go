package trust

import (
	"net/http"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/shared/apperr"
)

type Thresholds struct {
	Allow  int
	StepUp int
}

func Enforce(th Thresholds) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == http.MethodOptions {
				return next(c)
			}
			tc, _ := Get(c)
			if tc == nil {
				return apperr.New("FORBIDDEN", http.StatusForbidden, "Tidak punya akses.").
					WithField("reason", "trust_missing")
			}

			if tc.Score >= th.Allow {
				return next(c)
			}
			if tc.Score >= th.StepUp {
				return apperr.New("FORBIDDEN", http.StatusForbidden, "Step-up diperlukan.").
					WithField("reason", "stepup_required").
					WithField("trust_score", tc.Score).
					WithField("policy", tc.Policy)
			}
			return apperr.New("FORBIDDEN", http.StatusForbidden, "Tidak punya akses.").
				WithField("reason", "trust_low").
				WithField("trust_score", tc.Score).
				WithField("policy", tc.Policy)
		}
	}
}
