
."
".package http

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/usecase"
	"example.com/your-api/internal/shared/apperr"
	"example.com/your-api/internal/transport/http/presenter"
)

type ReleaseHandler struct {
	dash *usecase.Dashboard
}

func NewReleaseHandler(d *usecase.Dashboard) *ReleaseHandler { return &ReleaseHandler{dash: d} }

type ReleaseResponse struct {
	ReleaseID     string `json:"release_id"`
	SiteID        string `json:"site_id"`
	Status        string `json:"status"`
	SizeBytes     int64  `json:"size_bytes"`
	ErrorCode     string `json:"error_code"`
	CreatedAtUnix int64  `json:"created_at_unix"`
	UpdatedAtUnix int64  `json:"updated_at_unix"`
}

type ListReleasesResponse struct {
	Items []ReleaseResponse `json:"items"`
}

func (h *ReleaseHandler) ListBySite(c echo.Context) error {
	siteID := strings.TrimSpace(c.Param("site_id"))
	if siteID == "" {
		return apperr.New("BAD_REQUEST", http.StatusBadRequest, "Permintaan tidak valid.")
	}

	limit := 50
	if v := strings.TrimSpace(c.QueryParam("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	rels, err := h.dash.ListReleases(c.Request().Context(), domain.SiteID(siteID), limit)
	if err != nil {
		return err
	}

	items := make([]ReleaseResponse, 0, len(rels))
	for _, r := range rels {
		items = append(items, ReleaseResponse{
			ReleaseID:     r.ID.String(),
			SiteID:        r.SiteID.String(),
			Status:        string(r.Status),
			SizeBytes:     r.SizeBytes,
			ErrorCode:     r.ErrorCode,
			CreatedAtUnix: r.CreatedAt.UTC().Unix(),
			UpdatedAtUnix: r.UpdatedAt.UTC().Unix(),
		})
	}

	setNoStore(c)
	return presenter.OK(c, ListReleasesResponse{Items: items})
}

func (h *ReleaseHandler) GetBySite(c echo.Context) error {
	siteID := strings.TrimSpace(c.Param("site_id"))
	releaseID := strings.TrimSpace(c.Param("release_id"))
	if siteID == "" || releaseID == "" {
		return apperr.New("BAD_REQUEST", http.StatusBadRequest, "Permintaan tidak valid.")
	}

	r, err := h.dash.GetRelease(c.Request().Context(), domain.SiteID(siteID), domain.ReleaseID(releaseID))
	if err != nil {
		return mapReleaseErr(err)
	}

	setNoStore(c)
	return presenter.OK(c, ReleaseResponse{
		ReleaseID:     r.ID.String(),
		SiteID:        r.SiteID.String(),
		Status:        string(r.Status),
		SizeBytes:     r.SizeBytes,
		ErrorCode:     r.ErrorCode,
		CreatedAtUnix: r.CreatedAt.UTC().Unix(),
		UpdatedAtUnix: r.UpdatedAt.UTC().Unix(),
	})
}

func mapReleaseErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrReleaseNotFound):
		return apperr.New("RELEASE_NOT_FOUND", http.StatusNotFound, "Release tidak ditemukan.")
	default:
		return err
	}
}
