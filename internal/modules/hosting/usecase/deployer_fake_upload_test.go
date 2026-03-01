package usecase

import (
	"context"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
)

type fakeUploadStore struct {
	u           domain.Upload
	lastStatus  domain.UploadStatus
	lastSize    int64
	lastRelease domain.ReleaseID
}

func (f *fakeUploadStore) Put(context.Context, domain.Upload) error { return nil }
func (f *fakeUploadStore) Get(context.Context, domain.UploadID) (domain.Upload, error) {
	return f.u, nil
}
func (f *fakeUploadStore) UpdateStatus(_ context.Context, _ domain.UploadID, st domain.UploadStatus) error {
	f.lastStatus = st
	return nil
}
func (f *fakeUploadStore) UpdateStatusSizeAndReleaseID(_ context.Context, _ domain.UploadID, st domain.UploadStatus, sz int64, rid domain.ReleaseID) error {
	f.lastStatus, f.lastSize, f.lastRelease = st, sz, rid
	return nil
}

var _ ports.UploadStore = (*fakeUploadStore)(nil)
