//go:build component
// +build component

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

type refreshSpy struct{ got string }

func (s *refreshSpy) Refresh(ctx context.Context, in usecase.RefreshInput) (usecase.RefreshOutput, error) {
	s.got = in.RefreshToken
	return usecase.RefreshOutput{
		AccessToken: "jwt", AccessExp: time.Now().Add(30 * time.Minute),
		RefreshToken: "r2", RefreshExp: time.Now().Add(24 * time.Hour),
		CSRFToken: "c2", TrustLevel: "aal1",
	}, nil
}
func (s *refreshSpy) Logout(ctx context.Context, in usecase.LogoutInput) error { return nil }

func TestSessionHandler_RefreshSetsCookies(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = presenter.HTTPErrorHandler

	cfg := config.LoadAuth()
	cfg.Session.RefreshCookieName = "refresh"
	cfg.Session.CSRFCookieName = "csrf"
	cfg.Security.CookieSecure = false

	spy := &refreshSpy{}
	h := NewSessionHandler(spy, cfg)

	req := httptest.NewRequest(http.MethodPost, "/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh", Value: "r1"})
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Refresh(c); err != nil {
		t.Fatal(err)
	}
	if spy.got != "r1" {
		t.Fatalf("want refresh token r1 got %q", spy.got)
	}
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), `"access_token"`) {
		t.Fatalf("bad response code=%d body=%s", rec.Code, rec.Body.String())
	}
	if len(rec.Header().Values("Set-Cookie")) < 2 {
		t.Fatalf("expected cookies got %v", rec.Header().Values("Set-Cookie"))
	}
}
