package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/config"
)

func readCookie(c echo.Context, name string) string {
	ck, err := c.Cookie(name)
	if err != nil || ck == nil {
		return ""
	}
	return strings.TrimSpace(ck.Value)
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
