package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/config"
	"example.com/your-api/internal/modules/auth/usecase"
	"example.com/your-api/internal/transport/http/presenter"
)

type fakeSessFlow struct{}

func (f fakeSessFlow) Refresh(ctx context.Context, in usecase.RefreshInput) (usecase.RefreshOutput, error) {
	return usecase.RefreshOutput{
		AccessToken: "jwt", AccessExp: time.Now().Add(30 * time.Minute),
		RefreshToken: "r2", RefreshExp: time.Now().Add(24 * time.Hour),
		CSRFToken: "c2", TrustLevel: "aal1",
	}, nil
}
func (f fakeSessFlow) Logout(ctx context.Context, in usecase.LogoutInput) error { return nil }

func TestSessionRefresh_RequiresCSRF(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = presenter.HTTPErrorHandler

	cfg := config.LoadAuth()
	cfg.Session.RefreshCookieName = "refresh"
	cfg.Session.CSRFCookieName = "csrf"
	cfg.Security.CookieSecure = false

	h := NewSessionHandler(fakeSessFlow{}, cfg)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh", Value: "r"})
	req.AddCookie(&http.Cookie{Name: "csrf", Value: "c"}) // header intentionally missing
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Refresh(c); err != nil {
		e.HTTPErrorHandler(err, c)
	}
	if rec.Code != http.StatusForbidden {
		t.Fatalf("want 403 got %d body=%s", rec.Code, rec.Body.String())
	}
}

func TestSessionRefresh_SetsCookies(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = presenter.HTTPErrorHandler

	cfg := config.LoadAuth()
	cfg.Session.RefreshCookieName = "refresh"
	cfg.Session.CSRFCookieName = "csrf"
	cfg.Security.CookieSecure = false

	h := NewSessionHandler(fakeSessFlow{}, cfg)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh", Value: "r"})
	req.AddCookie(&http.Cookie{Name: "csrf", Value: "c"})
	req.Header.Set("X-CSRF-Token", "c")

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Refresh(c); err != nil {
		e.HTTPErrorHandler(err, c)
	}
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"access_token"`) {
		t.Fatalf("bad code=%d body=%s", rec.Code, rec.Body.String())
	}
	if len(rec.Header().Values("Set-Cookie")) < 2 {
		t.Fatalf("expected cookies got %v", rec.Header().Values("Set-Cookie"))
	}
}
