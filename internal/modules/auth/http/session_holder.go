package http

import (
	"net/http"
	"sync/atomic"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/shared/apperr"
)

var sessionHandler atomic.Value // *SessionHandler

func SetSessionHandler(h *SessionHandler) { sessionHandler.Store(h) }

func Refresh(c echo.Context) error {
	h, _ := sessionHandler.Load().(*SessionHandler)
	if h == nil {
		return apperr.New(domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	return h.Refresh(c)
}

func Logout(c echo.Context) error {
	h, _ := sessionHandler.Load().(*SessionHandler)
	if h == nil {
		return apperr.New(domain.ErrInternal, http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	return h.Logout(c)
}
