package usecase

import (
	"context"
	"time"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
)

type fakeSiteStore struct {
	lastSite    domain.SiteID
	lastRelease domain.ReleaseID
	byID        map[string]domain.Site
}

func (f *fakeSiteStore) Put(_ context.Context, s domain.Site) error {
	if f.byID == nil {
		f.byID = map[string]domain.Site{}
	}
	f.byID[s.ID.String()] = s
	return nil
}

func (f *fakeSiteStore) Get(_ context.Context, id domain.SiteID) (domain.Site, error) {
	if f.byID == nil {
		return domain.Site{}, domain.ErrSiteNotFound
	}
	s, ok := f.byID[id.String()]
	if !ok {
		return domain.Site{}, domain.ErrSiteNotFound
	}
	return s, nil
}

func (f *fakeSiteStore) List(_ context.Context, _ int) ([]domain.Site, error) {
	if f.byID == nil {
		return nil, nil
	}
	out := make([]domain.Site, 0, len(f.byID))
	for _, s := range f.byID {
		out = append(out, s)
	}
	return out, nil
}

func (f *fakeSiteStore) UpdateCurrentRelease(_ context.Context, id domain.SiteID, rid domain.ReleaseID) error {
	f.lastSite, f.lastRelease = id, rid
	if f.byID != nil {
		s, ok := f.byID[id.String()]
		if ok {
			s.CurrentRelease = rid
			s.Suspended = false
			s.UpdatedAt = time.Now().UTC()
			f.byID[id.String()] = s
		}
	}
	return nil
}

type fakeReleaseStore struct {
	lastStatus domain.ReleaseStatus
	lastSize   int64
	lastCode   string
	byID       map[string]domain.Release
}

func (f *fakeReleaseStore) Put(_ context.Context, r domain.Release) error {
	if f.byID == nil {
		f.byID = map[string]domain.Release{}
	}
	f.byID[r.ID.String()] = r
	return nil
}

func (f *fakeReleaseStore) Get(_ context.Context, id domain.ReleaseID) (domain.Release, error) {
	if f.byID == nil {
		return domain.Release{}, domain.ErrReleaseNotFound
	}
	r, ok := f.byID[id.String()]
	if !ok {
		return domain.Release{}, domain.ErrReleaseNotFound
	}
	return r, nil
}

func (f *fakeReleaseStore) ListBySite(_ context.Context, siteID domain.SiteID, _ int) ([]domain.Release, error) {
	if f.byID == nil {
		return nil, nil
	}
	out := make([]domain.Release, 0, len(f.byID))
	for _, r := range f.byID {
		if r.SiteID == siteID {
			out = append(out, r)
		}
	}
	return out, nil
}

func (f *fakeReleaseStore) UpdateStatus(_ context.Context, id domain.ReleaseID, st domain.ReleaseStatus, sz int64, code string) error {
	f.lastStatus, f.lastSize, f.lastCode = st, sz, code
	if f.byID != nil {
		r, ok := f.byID[id.String()]
		if ok {
			r.Status = st
			r.SizeBytes = sz
			r.ErrorCode = code
			r.UpdatedAt = time.Now().UTC()
			f.byID[id.String()] = r
		}
	}
	return nil
}

// Verifikasi interface implementation
var _ ports.SiteStore = (*fakeSiteStore)(nil)
var _ ports.ReleaseStore = (*fakeReleaseStore)(nil)
