package http

import (
	"net/http"
	"sync/atomic"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/shared/apperr"
)

var siteHandler atomic.Value // *SiteHandler

func SetSiteHandler(h *SiteHandler) { siteHandler.Store(h) }

func mustSiteHandler() *SiteHandler {
	h, _ := siteHandler.Load().(*SiteHandler)
	if h == nil {
		panic("hosting site handler not set")
	}
	return h
}

func ListSites(c echo.Context) error {
	h, _ := siteHandler.Load().(*SiteHandler)
	if h == nil {
		return apperr.New("INTERNAL", http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	return h.List(c)
}

func CreateSite(c echo.Context) error {
	h, _ := siteHandler.Load().(*SiteHandler)
	if h == nil {
		return apperr.New("INTERNAL", http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	return h.Create(c)
}

func GetSite(c echo.Context) error {
	h, _ := siteHandler.Load().(*SiteHandler)
	if h == nil {
		return apperr.New("INTERNAL", http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	return h.Get(c)
}
