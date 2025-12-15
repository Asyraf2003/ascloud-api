package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/shared/apperr"
)

type AAL string

const (
	AAL1 AAL = "aal1"
	AAL2 AAL = "aal2"
	AAL3 AAL = "aal3"
)

func RequireAAL(min AAL) echo.MiddlewareFunc {
	minRank := aalRank(string(min))

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			got, _ := AALLevel(c) // dari context helper, bukan c.Get("aal") liar
			if aalRank(got) < minRank {
				return apperr.New(domain.ErrForbidden, http.StatusForbidden, "Step-up diperlukan.").
					WithField("reason", "insufficient_aal").
					WithField("need", string(min)).
					WithField("got", got)
			}
			return next(c)
		}
	}
}

func aalRank(a string) int {
	switch strings.ToLower(a) {
	case "aal1":
		return 1
	case "aal2":
		return 2
	case "aal3":
		return 3
	default:
		return 0
	}
}
