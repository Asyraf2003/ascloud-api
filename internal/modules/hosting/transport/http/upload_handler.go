package http

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/usecase"
	"example.com/your-api/internal/shared/apperr"
	"example.com/your-api/internal/transport/http/presenter"
)

type UploadHandler struct {
	svc *usecase.Service
}

func NewUploadHandler(svc *usecase.Service) *UploadHandler { return &UploadHandler{svc: svc} }

func (h *UploadHandler) Initiate(c echo.Context) error {
	siteID := strings.TrimSpace(c.Param("site_id"))
	if siteID == "" {
		return apperr.New("BAD_REQUEST", http.StatusBadRequest, "Permintaan tidak valid.")
	}

	out, err := h.svc.InitiateUpload(c.Request().Context(), domain.SiteID(siteID))
	if err != nil {
		return mapHostingErr(err)
	}

	setNoStore(c)
	return presenter.Created(c, InitiateUploadResponse{
		UploadID:      out.UploadID.String(),
		ObjectKey:     out.ObjectKey,
		PutURL:        out.PutURL,
		ExpiresAtUnix: out.ExpiresAtUnix,
		MaxBytes:      out.MaxBytes,
	})
}

func (h *UploadHandler) Complete(c echo.Context) error {
	siteID := strings.TrimSpace(c.Param("site_id"))
	uploadID := strings.TrimSpace(c.Param("upload_id"))
	if siteID == "" || uploadID == "" {
		return apperr.New("BAD_REQUEST", http.StatusBadRequest, "Permintaan tidak valid.")
	}

	out, err := h.svc.CompleteUpload(c.Request().Context(), domain.SiteID(siteID), domain.UploadID(uploadID))
	if err != nil {
		return mapHostingErr(err)
	}

	setNoStore(c)
	return presenter.OK(c, CompleteUploadResponse{
		Status:    string(out.Status),
		SizeBytes: out.Size,
	})
}

func mapHostingErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrUploadNotFound):
		return apperr.New("UPLOAD_NOT_FOUND", http.StatusNotFound, "Upload tidak ditemukan.")
	case errors.Is(err, domain.ErrUploadTooLarge):
		return apperr.New("UPLOAD_TOO_LARGE", http.StatusRequestEntityTooLarge, "Upload terlalu besar.")
	case errors.Is(err, domain.ErrUploadSiteMismatch):
		return apperr.New("UPLOAD_SITE_MISMATCH", http.StatusBadRequest, "Upload tidak sesuai site.")
	case errors.Is(err, domain.ErrUploadAlreadyQueued):
		return apperr.New("UPLOAD_ALREADY_QUEUED", http.StatusConflict, "Upload sudah diantrikan.")
	default:
		return err
	}
}

func setNoStore(c echo.Context) {
	c.Response().Header().Set("Cache-Control", "no-store")
	c.Response().Header().Set("Pragma", "no-cache")
	c.Response().Header().Set("Expires", time.Unix(0, 0).UTC().Format(time.RFC1123))
}
