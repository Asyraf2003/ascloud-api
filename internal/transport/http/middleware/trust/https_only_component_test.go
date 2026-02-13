//go:build component
// +build component

package trust

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/transport/http/presenter"
)

func TestRequireHTTPS_BlocksHTTP(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = presenter.HTTPErrorHandler

	e.GET("/p", func(c echo.Context) error { return c.String(http.StatusOK, "ok") }, RequireHTTPS())

	req := httptest.NewRequest(http.MethodGet, "/p", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected %d got %d body=%s", http.StatusForbidden, rec.Code, rec.Body.String())
	}
}

func TestRequireHTTPS_AllowsForwardedProtoHTTPS(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = presenter.HTTPErrorHandler

	e.GET("/p", func(c echo.Context) error { return c.String(http.StatusOK, "ok") }, RequireHTTPS())

	req := httptest.NewRequest(http.MethodGet, "/p", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
}

func TestRequireHTTPS_AllowsTLS(t *testing.T) {
	e := echo.New()
	e.HTTPErrorHandler = presenter.HTTPErrorHandler

	e.GET("/p", func(c echo.Context) error { return c.String(http.StatusOK, "ok") }, RequireHTTPS())

	req := httptest.NewRequest(http.MethodGet, "/p", nil)
	req.TLS = &tls.ConnectionState{}
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected %d got %d body=%s", http.StatusOK, rec.Code, rec.Body.String())
	}
}
