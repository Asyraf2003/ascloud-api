// TODO justify: >100 lines sementara; akan dipecah saat milestone 9.

package usecase

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
)

type DashboardConfig struct {
	BaseDomain string // contoh: asyrafcloud.my.id
}

type Dashboard struct {
	cfg   DashboardConfig
	sites ports.SiteStore
	rel   ports.ReleaseStore
	edge  ports.EdgeStore
}

func NewDashboard(cfg DashboardConfig, sites ports.SiteStore, rel ports.ReleaseStore, edge ports.EdgeStore) *Dashboard {
	return &Dashboard{cfg: cfg, sites: sites, rel: rel, edge: edge}
}

func (d *Dashboard) CreateSite(ctx context.Context, requestedSiteID string) (domain.Site, error) {
	sid := normalizeSiteID(requestedSiteID)
	if sid == "" {
		sid = autoSiteID()
	}
	if sid == "" || d.cfg.BaseDomain == "" {
		return domain.Site{}, domain.ErrSiteInvalid
	}

	host := sid + "." + strings.TrimSpace(d.cfg.BaseDomain)
	now := time.Now().UTC()

	site := domain.Site{
		ID:             domain.SiteID(sid),
		Host:           host,
		Suspended:      true,
		CurrentRelease: "",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := d.sites.Put(ctx, site); err != nil {
		return domain.Site{}, err
	}

	// create initial mapping: suspended=true (before first activation)
	if d.edge != nil {
		_ = d.edge.PutHostMapping(ctx, site.Host, site.ID, "", true)
	}
	return site, nil
}

func (d *Dashboard) ListSites(ctx context.Context, limit int) ([]domain.Site, error) {
	sites, err := d.sites.List(ctx, limit)
	if err != nil {
		return nil, err
	}
	sort.Slice(sites, func(i, j int) bool { return sites[i].CreatedAt.After(sites[j].CreatedAt) })
	return sites, nil
}

func (d *Dashboard) GetSite(ctx context.Context, siteID domain.SiteID) (domain.Site, error) {
	return d.sites.Get(ctx, siteID)
}

func (d *Dashboard) ListReleases(ctx context.Context, siteID domain.SiteID, limit int) ([]domain.Release, error) {
	rels, err := d.rel.ListBySite(ctx, siteID, limit)
	if err != nil {
		return nil, err
	}
	sort.Slice(rels, func(i, j int) bool { return rels[i].CreatedAt.After(rels[j].CreatedAt) })
	return rels, nil
}

func (d *Dashboard) GetRelease(ctx context.Context, siteID domain.SiteID, releaseID domain.ReleaseID) (domain.Release, error) {
	r, err := d.rel.Get(ctx, releaseID)
	if err != nil {
		return domain.Release{}, err
	}
	if strings.TrimSpace(r.SiteID.String()) != strings.TrimSpace(siteID.String()) {
		return domain.Release{}, domain.ErrReleaseNotFound
	}
	return r, nil
}

func (d *Dashboard) Rollback(ctx context.Context, siteID domain.SiteID, releaseID domain.ReleaseID) (domain.Site, error) {
	site, err := d.sites.Get(ctx, siteID)
	if err != nil {
		return domain.Site{}, err
	}

	r, err := d.rel.Get(ctx, releaseID)
	if err != nil {
		return domain.Site{}, err
	}
	if strings.TrimSpace(r.SiteID.String()) != strings.TrimSpace(siteID.String()) {
		return domain.Site{}, domain.ErrReleaseNotFound
	}

	// Edge first (traffic switch), then metadata DB.
	if d.edge != nil {
		if err := d.edge.PutHostMapping(ctx, site.Host, site.ID, r.ID, false); err != nil {
			return domain.Site{}, err
		}
	}

	if err := d.sites.UpdateCurrentRelease(ctx, siteID, r.ID); err != nil {
		return domain.Site{}, err
	}

	site.CurrentRelease = r.ID
	site.Suspended = false
	site.UpdatedAt = time.Now().UTC()
	return site, nil
}

func normalizeSiteID(v string) string {
	s := strings.TrimSpace(strings.ToLower(v))
	if s == "" {
		return ""
	}
	// basic DNS label allowlist: [a-z0-9-], no leading/trailing '-'
	for i := 0; i < len(s); i++ {
		c := s[i]
		ok := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-'
		if !ok {
			return ""
		}
	}
	if strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") {
		return ""
	}
	if len(s) < 3 || len(s) > 40 {
		return ""
	}
	return s
}

func autoSiteID() string {
	// s- + 8 hex = aman untuk DNS label
	u := uuid.NewString()
	if len(u) < 8 {
		return ""
	}
	return "s-" + u[:8]
}
