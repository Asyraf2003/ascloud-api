package http

import (
	"net/http"
	"sync/atomic"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/shared/apperr"
)

var googleHandler atomic.Value // *GoogleHandler

func SetGoogleHandler(h *GoogleHandler) { googleHandler.Store(h) }

func GoogleStart(c echo.Context) error {
	h, _ := googleHandler.Load().(*GoogleHandler)
	if h == nil {
		return apperr.New(domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	return h.Start(c)
}

func GoogleCallback(c echo.Context) error {
	h, _ := googleHandler.Load().(*GoogleHandler)
	if h == nil {
		return apperr.New(domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	return h.Callback(c)
}
