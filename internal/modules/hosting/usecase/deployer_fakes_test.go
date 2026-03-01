package usecase

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

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

type fakeObjectStore struct {
	zipBytes   []byte
	putKeys    []string
	deleteKey  string
	putFailErr error
}

func (f *fakeObjectStore) PresignPutZip(context.Context, string, int64, time.Duration) (string, error) {
	return "", nil
}
func (f *fakeObjectStore) Head(context.Context, string) (int64, error) { return 0, nil }
func (f *fakeObjectStore) Get(context.Context, string) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(f.zipBytes)), nil
}
func (f *fakeObjectStore) Put(_ context.Context, key string, body io.Reader, _ string, _ string) error {
	if f.putFailErr != nil {
		return f.putFailErr
	}
	_, _ = io.ReadAll(body)
	f.putKeys = append(f.putKeys, key)
	return nil
}
func (f *fakeObjectStore) Delete(_ context.Context, key string) error {
	f.deleteKey = key
	return nil
}

func makeZipBytes(t *testing.T, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for name, b := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(b); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func tempTmpDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "tmp")
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

var _ ports.UploadStore = (*fakeUploadStore)(nil)
var _ ports.SiteStore = (*fakeSiteStore)(nil)
var _ ports.ReleaseStore = (*fakeReleaseStore)(nil)
var _ ports.ObjectStore = (*fakeObjectStore)(nil)
