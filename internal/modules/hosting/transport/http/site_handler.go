package http

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

type SiteHandler struct {
	dash *usecase.Dashboard
}

func NewSiteHandler(d *usecase.Dashboard) *SiteHandler { return &SiteHandler{dash: d} }

type CreateSiteRequest struct {
	SiteID string `json:"site_id"`
}

type SiteResponse struct {
	SiteID         string `json:"site_id"`
	Host           string `json:"host"`
	Suspended      bool   `json:"suspended"`
	CurrentRelease string `json:"current_release_id"`
	CreatedAtUnix  int64  `json:"created_at_unix"`
	UpdatedAtUnix  int64  `json:"updated_at_unix"`
}

type ListSitesResponse struct {
	Items []SiteResponse `json:"items"`
}

func (h *SiteHandler) Create(c echo.Context) error {
	var req CreateSiteRequest
	_ = c.Bind(&req)

	out, err := h.dash.CreateSite(c.Request().Context(), req.SiteID)
	if err != nil {
		return mapSiteErr(err)
	}

	setNoStore(c)
	return presenter.Created(c, SiteResponse{
		SiteID:         out.ID.String(),
		Host:           out.Host,
		Suspended:      out.Suspended,
		CurrentRelease: out.CurrentRelease.String(),
		CreatedAtUnix:  out.CreatedAt.UTC().Unix(),
		UpdatedAtUnix:  out.UpdatedAt.UTC().Unix(),
	})
}

func (h *SiteHandler) List(c echo.Context) error {
	limit := 50
	if v := strings.TrimSpace(c.QueryParam("limit")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	sites, err := h.dash.ListSites(c.Request().Context(), limit)
	if err != nil {
		return err
	}

	items := make([]SiteResponse, 0, len(sites))
	for _, s := range sites {
		items = append(items, SiteResponse{
			SiteID:         s.ID.String(),
			Host:           s.Host,
			Suspended:      s.Suspended,
			CurrentRelease: s.CurrentRelease.String(),
			CreatedAtUnix:  s.CreatedAt.UTC().Unix(),
			UpdatedAtUnix:  s.UpdatedAt.UTC().Unix(),
		})
	}

	setNoStore(c)
	return presenter.OK(c, ListSitesResponse{Items: items})
}

func (h *SiteHandler) Get(c echo.Context) error {
	siteID := strings.TrimSpace(c.Param("site_id"))
	if siteID == "" {
		return apperr.New("BAD_REQUEST", http.StatusBadRequest, "Permintaan tidak valid.")
	}

	s, err := h.dash.GetSite(c.Request().Context(), domain.SiteID(siteID))
	if err != nil {
		return mapSiteErr(err)
	}

	setNoStore(c)
	return presenter.OK(c, SiteResponse{
		SiteID:         s.ID.String(),
		Host:           s.Host,
		Suspended:      s.Suspended,
		CurrentRelease: s.CurrentRelease.String(),
		CreatedAtUnix:  s.CreatedAt.UTC().Unix(),
		UpdatedAtUnix:  s.UpdatedAt.UTC().Unix(),
	})
}

func mapSiteErr(err error) error {
	switch {
	case errors.Is(err, domain.ErrSiteNotFound):
		return apperr.New("SITE_NOT_FOUND", http.StatusNotFound, "Site tidak ditemukan.")
	case errors.Is(err, domain.ErrSiteAlreadyExists):
		return apperr.New("SITE_ALREADY_EXISTS", http.StatusConflict, "Site sudah ada.")
	default:
		return err
	}
}
