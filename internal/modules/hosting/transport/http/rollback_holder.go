package http

import (
	"net/http"
	"sync/atomic"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/shared/apperr"
)

var rollbackHandler atomic.Value // *RollbackHandler

func SetRollbackHandler(h *RollbackHandler) { rollbackHandler.Store(h) }

func RollbackSite(c echo.Context) error {
	h, _ := rollbackHandler.Load().(*RollbackHandler)
	if h == nil {
		return apperr.New("INTERNAL", http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	return h.Rollback(c)
}
