package usecase

import (
	"context"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
)

type fakeSiteStore struct {
	lastSite    domain.SiteID
	lastRelease domain.ReleaseID
}

func (f *fakeSiteStore) Get(context.Context, domain.SiteID) (domain.Site, error) {
	return domain.Site{}, nil
}
func (f *fakeSiteStore) UpdateCurrentRelease(_ context.Context, id domain.SiteID, rid domain.ReleaseID) error {
	f.lastSite, f.lastRelease = id, rid
	return nil
}

type fakeReleaseStore struct {
	lastStatus domain.ReleaseStatus
	lastSize   int64
	lastCode   string
}

func (f *fakeReleaseStore) Put(context.Context, domain.Release) error { return nil }
func (f *fakeReleaseStore) UpdateStatus(_ context.Context, _ domain.ReleaseID, st domain.ReleaseStatus, sz int64, code string) error {
	f.lastStatus, f.lastSize, f.lastCode = st, sz, code
	return nil
}

var _ ports.SiteStore = (*fakeSiteStore)(nil)
var _ ports.ReleaseStore = (*fakeReleaseStore)(nil)
