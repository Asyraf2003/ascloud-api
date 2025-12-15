package http

import (
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/config"
)

func setAuthCookies(c echo.Context, cfg config.AuthConfig, refresh string, refreshExp time.Time, csrf string) error {
	now := time.Now()
	maxAge := int(refreshExp.Sub(now).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}

	ss := parseSameSite(cfg.Security.CookieSameSite)
	path := "/v1/auth"

	refreshC := &http.Cookie{
		Name:     cfg.Session.RefreshCookieName,
		Value:    refresh,
		Path:     path,
		Domain:   cfg.Security.CookieDomain,
		Secure:   cfg.Security.CookieSecure,
		HttpOnly: true,
		SameSite: ss,
		MaxAge:   maxAge,
		Expires:  refreshExp,
	}

	csrfC := &http.Cookie{
		Name:     cfg.Session.CSRFCookieName,
		Value:    csrf,
		Path:     path,
		Domain:   cfg.Security.CookieDomain,
		Secure:   cfg.Security.CookieSecure,
		HttpOnly: false,
		SameSite: ss,
		MaxAge:   maxAge,
		Expires:  refreshExp,
	}

	c.SetCookie(refreshC)
	c.SetCookie(csrfC)
	return nil
}

func parseSameSite(v string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case "lax":
		return http.SameSiteLaxMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteStrictMode
	}
}
