package http

import (
	"net/http"
	"sync/atomic"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/shared/apperr"
)

var uploadHandler atomic.Value // *UploadHandler

func SetUploadHandler(h *UploadHandler) { uploadHandler.Store(h) }

func InitiateUpload(c echo.Context) error {
	h, _ := uploadHandler.Load().(*UploadHandler)
	if h == nil {
		return apperr.New("INTERNAL", http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	return h.Initiate(c)
}

func CompleteUpload(c echo.Context) error {
	h, _ := uploadHandler.Load().(*UploadHandler)
	if h == nil {
		return apperr.New("INTERNAL", http.StatusInternalServerError, "Terjadi kesalahan.")
	}
	return h.Complete(c)
}
