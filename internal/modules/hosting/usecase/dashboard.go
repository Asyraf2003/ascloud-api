// TODO justify: >100 lines sementara; akan dipecah saat milestone 9.

package usecase

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
	"example.com/your-api/internal/shared/apperr"
	"example.com/your-api/internal/shared/requestid"
)

type DashboardConfig struct {
	BaseDomain string
}

type Dashboard struct {
	cfg   DashboardConfig
	sites ports.SiteStore
	rel   ports.ReleaseStore
	edge  ports.EdgeStore
	audit ports.AuditSink
}

func NewDashboard(cfg DashboardConfig, sites ports.SiteStore, rel ports.ReleaseStore, edge ports.EdgeStore, audit ports.AuditSink) *Dashboard {
	return &Dashboard{cfg: cfg, sites: sites, rel: rel, edge: edge, audit: audit}
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

	if d.edge != nil {
		_ = d.edge.PutHostMapping(ctx, site.Host, site.ID, "", true)
	}

	d.audit0(ctx, site.ID, "hosting_site_created", map[string]any{
		"request_id": rid(ctx),
		"host":       site.Host,
		"suspended":  site.Suspended,
	})

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
	d.audit0(ctx, siteID, "hosting_rollback_requested", map[string]any{
		"request_id": rid(ctx),
		"release_id": releaseID,
	})

	site, err := d.sites.Get(ctx, siteID)
	if err != nil {
		d.audit_fail(ctx, siteID, releaseID, "get_site", err)
		return domain.Site{}, err
	}

	r, err := d.rel.Get(ctx, releaseID)
	if err != nil {
		d.audit_fail(ctx, siteID, releaseID, "get_release", err)
		return domain.Site{}, err
	}
	if strings.TrimSpace(r.SiteID.String()) != strings.TrimSpace(siteID.String()) {
		d.audit_fail(ctx, siteID, releaseID, "validate_release_site", domain.ErrReleaseNotFound)
		return domain.Site{}, domain.ErrReleaseNotFound
	}

	if d.edge != nil {
		if err := d.edge.PutHostMapping(ctx, site.Host, site.ID, r.ID, false); err != nil {
			d.audit_fail(ctx, siteID, releaseID, "edge_put_host_mapping", err)
			return domain.Site{}, err
		}
	}

	if err := d.sites.UpdateCurrentRelease(ctx, siteID, r.ID); err != nil {
		d.audit_fail(ctx, siteID, releaseID, "ddb_update_current_release", err)
		return domain.Site{}, err
	}

	site.CurrentRelease = r.ID
	site.Suspended = false
	site.UpdatedAt = time.Now().UTC()

	d.audit0(ctx, siteID, "hosting_rollback_succeeded", map[string]any{
		"request_id":         rid(ctx),
		"release_id":         releaseID,
		"current_release_id": r.ID,
		"host":               site.Host,
	})

	return site, nil
}

// Helpers
func (d *Dashboard) audit0(ctx context.Context, sid domain.SiteID, event string, meta map[string]any) {
	if d.audit == nil || strings.TrimSpace(sid.String()) == "" {
		return
	}
	_ = d.audit.Record(ctx, ports.AuditEvent{SiteID: sid, Event: event, At: time.Now(), Meta: meta})
}

func (d *Dashboard) audit_fail(ctx context.Context, sid domain.SiteID, rid_ domain.ReleaseID, stage string, err error) {
	d.audit0(ctx, sid, "hosting_rollback_failed", map[string]any{
		"request_id": rid(ctx),
		"release_id": rid_,
		"stage":      stage,
		"error_code": auditCode(err),
	})
}

func rid(ctx context.Context) string {
	id, _ := requestid.From(ctx)
	return id
}

func auditCode(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, domain.ErrReleaseNotFound) {
		return "hosting.release_not_found"
	}
	if errors.Is(err, domain.ErrSiteNotFound) {
		return "hosting.site_not_found"
	}
	if ae, ok := apperr.As(err); ok {
		return ae.Code
	}
	return "hosting.internal_error"
}

func normalizeSiteID(v string) string {
	s := strings.TrimSpace(strings.ToLower(v))
	if len(s) < 3 || len(s) > 40 {
		return ""
	}
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return ""
		}
	}
	if strings.HasPrefix(s, "-") || strings.HasSuffix(s, "-") {
		return ""
	}
	return s
}

func autoSiteID() string {
	u := uuid.NewString()
	if len(u) < 8 {
		return ""
	}
	return "s-" + u[:8]
}
