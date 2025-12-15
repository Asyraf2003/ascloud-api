package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/http/presenter"
	"example.com/your-api/internal/modules/auth/ports"
	j "example.com/your-api/internal/platform/token/jwt"
)

func TestJWTAuth_OK(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = presenter.HTTPErrorHandler

	iss, _ := j.NewHMACIssuer("iss", "aud", "kid", "secretsecret", 30*time.Minute)
	vf, _ := j.NewHMACVerifier("iss", "aud", "secretsecret")
	tok, _, _ := iss.IssueAccessToken(context.Background(), ports.AccessTokenRequest{AccountID: "a", SessionID: "s", TrustLevel: "aal1"})

	e.GET("/p", func(c echo.Context) error { return c.String(200, "ok") }, JWTAuth(vf))

	r := httptest.NewRequest(http.MethodGet, "/p", nil)
	r.Header.Set("Authorization", "Bearer "+tok)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)
	if w.Code != 200 {
		t.Fatalf("want 200 got %d body=%s", w.Code, w.Body.String())
	}
}

func TestJWTAuth_Missing(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = presenter.HTTPErrorHandler
	vf, _ := j.NewHMACVerifier("iss", "aud", "secretsecret")

	e.GET("/p", func(c echo.Context) error { return c.String(200, "ok") }, JWTAuth(vf))

	r := httptest.NewRequest(http.MethodGet, "/p", nil)
	w := httptest.NewRecorder()
	e.ServeHTTP(w, r)
	if w.Code != 401 || !strings.Contains(w.Body.String(), `"code":"AUTH_UNAUTHORIZED"`) {
		t.Fatalf("want 401 AUTH_UNAUTHORIZED got %d body=%s", w.Code, w.Body.String())
	}
}
