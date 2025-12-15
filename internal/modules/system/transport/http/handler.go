package http

import (
	"database/sql"
	"net/http"

	"github.com/labstack/echo/v4"
)

type Handler struct{}

func NewHandler() *Handler { return &Handler{} }

// Health: selalu OK kalau service jalan.
func (h *Handler) Health(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}

// Ready: OK kalau DB connect (kalau DB ada).
// Kalau DB nil, tetap OK (fase awal belum pakai DB).
func (h *Handler) Ready(c echo.Context) error {
	v := c.Get("db")
	db, _ := v.(*sql.DB)
	if db == nil {
		return c.JSON(http.StatusOK, map[string]any{"ready": true, "db": "skipped"})
	}
	if err := db.PingContext(c.Request().Context()); err != nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]any{"ready": false, "db": "down"})
	}
	return c.JSON(http.StatusOK, map[string]any{"ready": true, "db": "up"})
}
