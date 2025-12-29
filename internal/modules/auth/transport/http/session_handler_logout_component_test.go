//go:build component
// +build component

package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/config"
	"example.com/your-api/internal/modules/auth/usecase"
	"example.com/your-api/internal/transport/http/presenter"
)

type logoutSpy struct {
	called bool
	got    string
}

func (s *logoutSpy) Refresh(ctx context.Context, in usecase.RefreshInput) (usecase.RefreshOutput, error) {
	return usecase.RefreshOutput{}, nil
}
func (s *logoutSpy) Logout(ctx context.Context, in usecase.LogoutInput) error {
	s.called = true
	s.got = in.RefreshToken
	return nil
}

func TestSessionHandler_LogoutClearsCookies(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = presenter.HTTPErrorHandler

	cfg := config.LoadAuth()
	cfg.Session.RefreshCookieName = "refresh"
	cfg.Session.CSRFCookieName = "csrf"
	cfg.Security.CookieSecure = false

	spy := &logoutSpy{}
	h := NewSessionHandler(spy, cfg)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "refresh", Value: "r1"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Logout(c); err != nil {
		t.Fatal(err)
	}
	if !spy.called || spy.got != "r1" {
		t.Fatalf("logout not called correctly called=%v got=%q", spy.called, spy.got)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204 got %d", rec.Code)
	}
	cookies := strings.Join(rec.Header().Values("Set-Cookie"), "; ")
	if !strings.Contains(cookies, "refresh=") || !strings.Contains(cookies, "csrf=") {
		t.Fatalf("expected refresh+csrf cookies got %v", rec.Header().Values("Set-Cookie"))
	}
}

func TestSessionHandler_OptionsNoContent(t *testing.T) {
	e := echo.New()
	h := NewSessionHandler(&logoutSpy{}, config.LoadAuth())

	req := httptest.NewRequest(http.MethodOptions, "/v1/auth/logout", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Logout(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("want 204 got %d", rec.Code)
	}
}
