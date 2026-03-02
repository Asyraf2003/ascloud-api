package http

import (
	"errors"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/usecase"
	"example.com/your-api/internal/shared/apperr"
	"example.com/your-api/internal/transport/http/presenter"
)

type RollbackHandler struct {
	dash *usecase.Dashboard
}

func NewRollbackHandler(d *usecase.Dashboard) *RollbackHandler { return &RollbackHandler{dash: d} }

type RollbackRequest struct {
	ReleaseID string `json:"release_id"`
}

type RollbackResponse struct {
	SiteID         string `json:"site_id"`
	Host           string `json:"host"`
	CurrentRelease string `json:"current_release_id"`
	Suspended      bool   `json:"suspended"`
	UpdatedAtUnix  int64  `json:"updated_at_unix"`
}

func (h *RollbackHandler) Rollback(c echo.Context) error {
	siteID := strings.TrimSpace(c.Param("site_id"))
	if siteID == "" {
		return apperr.New("BAD_REQUEST", http.StatusBadRequest, "Permintaan tidak valid.")
	}

	var req RollbackRequest
	if err := c.Bind(&req); err != nil {
		return apperr.New("BAD_REQUEST", http.StatusBadRequest, "Permintaan tidak valid.")
	}
	rid := strings.TrimSpace(req.ReleaseID)
	if rid == "" {
		return apperr.New("BAD_REQUEST", http.StatusBadRequest, "Permintaan tidak valid.")
	}

	site, err := h.dash.Rollback(c.Request().Context(), domain.SiteID(siteID), domain.ReleaseID(rid))
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrReleaseNotFound):
			return apperr.New("RELEASE_NOT_FOUND", http.StatusNotFound, "Release tidak ditemukan.")
		default:
			return err
		}
	}

	setNoStore(c)
	return presenter.OK(c, RollbackResponse{
		SiteID:         site.ID.String(),
		Host:           site.Host,
		CurrentRelease: site.CurrentRelease.String(),
		Suspended:      site.Suspended,
		UpdatedAtUnix:  site.UpdatedAt.UTC().Unix(),
	})
}
