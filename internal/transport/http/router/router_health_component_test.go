//go:build component
// +build component

package router

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	jwt "example.com/your-api/internal/platform/token/jwt"
)

func mustVerifier(t *testing.T) *jwt.Verifier {
	t.Helper()
	v, err := jwt.NewHMACVerifier("iss", "aud", strings.Repeat("x", 32))
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestRouterHealth(t *testing.T) {
	e := echo.New()
	Register(e, mustVerifier(t))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}
