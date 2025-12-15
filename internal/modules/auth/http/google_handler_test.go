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
	"example.com/your-api/internal/http/presenter"
	"example.com/your-api/internal/modules/auth/usecase"
)

type fakeFlow struct{}

func (f fakeFlow) GoogleStart(ctx context.Context, in usecase.GoogleStartInput) (usecase.GoogleStartOutput, error) {
	return usecase.GoogleStartOutput{RedirectTo: "https://x", State: "s"}, nil
}
func (f fakeFlow) GoogleCallback(ctx context.Context, in usecase.GoogleCallbackInput) (usecase.GoogleCallbackOutput, error) {
	return usecase.GoogleCallbackOutput{
		AccessToken: "jwt", AccessExp: time.Now().Add(30 * time.Minute),
		RefreshToken: "r", RefreshExp: time.Now().Add(24 * time.Hour),
		CSRFToken: "c", TrustLevel: "aal1",
	}, nil
}

func TestGoogleHandler_StartRedirect(t *testing.T) {
	e := echo.New()
	h := NewGoogleHandler(fakeFlow{}, config.LoadAuth())
	req := httptest.NewRequest(http.MethodGet, "/v1/auth/google/start", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Start(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusFound || rec.Header().Get("Location") == "" {
		t.Fatalf("want 302 got %d loc=%s", rec.Code, rec.Header().Get("Location"))
	}
}

func TestGoogleHandler_CallbackSetsCookies(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = presenter.HTTPErrorHandler

	cfg := config.LoadAuth()
	cfg.Session.RefreshCookieName = "refresh"
	cfg.Session.CSRFCookieName = "csrf"
	cfg.Security.CookieSecure = false

	h := NewGoogleHandler(fakeFlow{}, cfg)

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/google/callback?code=x&state=y", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if err := h.Callback(c); err != nil {
		t.Fatal(err)
	}
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"access_token"`) {
		t.Fatalf("bad response code=%d body=%s", rec.Code, rec.Body.String())
	}
	got := rec.Header().Values("Set-Cookie")
	if len(got) < 2 {
		t.Fatalf("expected cookies got %v", got)
	}
}
