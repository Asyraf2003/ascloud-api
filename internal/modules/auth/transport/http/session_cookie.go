package http

import (
	"crypto/subtle"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/config"
	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/shared/apperr"
)

const csrfHeaderName = "X-CSRF-Token"

func readCookie(c echo.Context, name string) string {
	ck, err := c.Cookie(name)
	if err != nil || ck == nil {
		return ""
	}
	return strings.TrimSpace(ck.Value)
}

func requireCSRF(c echo.Context, cfg config.AuthConfig) error {
	c1 := readCookie(c, cfg.Session.CSRFCookieName)
	c2 := strings.TrimSpace(c.Request().Header.Get(csrfHeaderName))
	if c1 == "" || c2 == "" || subtle.ConstantTimeCompare([]byte(c1), []byte(c2)) != 1 {
		return apperr.New(domain.ErrCSRFInvalid, http.StatusForbidden, "Tidak punya akses.")
	}
	return nil
}

func clearAuthCookies(c echo.Context, cfg config.AuthConfig) {
	exp := time.Unix(0, 0).UTC()
	path := "/v1/auth"
	ss := parseSameSite(cfg.Security.CookieSameSite)

	c.SetCookie(&http.Cookie{
		Name: cfg.Session.RefreshCookieName, Value: "", Path: path,
		Domain: cfg.Security.CookieDomain, Secure: cfg.Security.CookieSecure,
		HttpOnly: true, SameSite: ss, MaxAge: -1, Expires: exp,
	})
	c.SetCookie(&http.Cookie{
		Name: cfg.Session.CSRFCookieName, Value: "", Path: path,
		Domain: cfg.Security.CookieDomain, Secure: cfg.Security.CookieSecure,
		HttpOnly: false, SameSite: ss, MaxAge: -1, Expires: exp,
	})
}
