//go:build component
// +build component

package middleware_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"

	jwt "example.com/your-api/internal/platform/token/jwt"
	"example.com/your-api/internal/transport/http/middleware"
	"example.com/your-api/internal/transport/http/router"
)

func mustVerifier(t *testing.T) *jwt.Verifier {
	t.Helper()
	v, err := jwt.NewHMACVerifier("iss", "aud", strings.Repeat("x", 32))
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestAccessLogAddsRequestID(t *testing.T) {
	log := slog.New(slog.NewJSONHandler(io.Discard, nil))

	e := echo.New()
	e.Use(middleware.RequestID())
	e.Use(middleware.AccessLog(log))

	// router.Register sekarang butuh verifier (buat protected route wiring).
	router.Register(e, mustVerifier(t))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatal("missing X-Request-Id header")
	}
}
