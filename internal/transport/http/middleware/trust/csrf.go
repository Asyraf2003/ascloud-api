package trust

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/shared/apperr"
)

func RequireCSRF(cookieName, headerName string) echo.MiddlewareFunc {
	return RequireCSRFWithCode(cookieName, headerName, "FORBIDDEN")
}

func RequireCSRFWithCode(cookieName, headerName, code string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == http.MethodOptions {
				return next(c)
			}
			ck, err := c.Cookie(cookieName)
			if err != nil || ck == nil || strings.TrimSpace(ck.Value) == "" {
				return apperr.New(code, http.StatusForbidden, "Tidak punya akses.").
					WithField("reason", "csrf_cookie_missing")
			}
			h := strings.TrimSpace(c.Request().Header.Get(headerName))
			if h == "" || subtle.ConstantTimeCompare([]byte(ck.Value), []byte(h)) != 1 {
				return apperr.New(code, http.StatusForbidden, "Tidak punya akses.").
					WithField("reason", "csrf_invalid")
			}
			Add(c, 15, "csrf_ok")
			return next(c)
		}
	}
}
