package usecase

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"example.com/your-api/internal/modules/hosting/ports"
)

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

var _ ports.ObjectStore = (*fakeObjectStore)(nil)
