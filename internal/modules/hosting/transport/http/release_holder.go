package http

import (
	"net/http"
	"sync/atomic"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/shared/apperr"
)

var releaseHandler atomic.Value // *ReleaseHandler

func SetReleaseHandler(h *ReleaseHandler) { releaseHandler.Store(h) }

func ListReleases(c echo.Context) error {
	h, _ := releaseHandler.Load().(*ReleaseHandler)
	if h == nil {
		return apperr.New("INTERNAL", http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	return h.ListBySite(c)
}

func GetRelease(c echo.Context) error {
	h, _ := releaseHandler.Load().(*ReleaseHandler)
	if h == nil {
		return apperr.New("INTERNAL", http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	return h.GetBySite(c)
}
