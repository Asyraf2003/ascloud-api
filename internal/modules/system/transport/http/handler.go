package http

import (
	"context"
	"net/http"
	"reflect"

	"github.com/labstack/echo/v4"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

// Health: selalu OK kalau service jalan.
func (h *Handler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}

// Ready: OK kalau DB connect (kalau DB ada).
// Kalau DB tidak ada / tidak support ping, tetap OK (fase awal belum pakai DB).
func (h *Handler) Ready(c echo.Context) error {
	v := c.Get("db")
	if v == nil {
		return c.JSON(http.StatusOK, map[string]any{"ready": true, "db": "skipped"})
	}

	p, ok := v.(interface {
		PingContext(context.Context) error
	})
	if !ok || isNil(p) {
		return c.JSON(http.StatusOK, map[string]any{"ready": true, "db": "skipped"})
	}

	if err := p.PingContext(c.Request().Context()); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]any{"ready": false, "db": "down"})
	}
	return c.JSON(http.StatusOK, map[string]any{"ready": true, "db": "up"})
}

func isNil(v any) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Ptr, reflect.Map, reflect.Slice, reflect.Func, reflect.Chan:
		return rv.IsNil()
	default:
		return false
	}
}
