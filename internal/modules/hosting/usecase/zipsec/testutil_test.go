package zipsec

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type zeroReader struct{}

func (zeroReader) Read(p []byte) (int, error) {
	for i := range p {
		p[i] = 0
	}
	return len(p), nil
}

func writeZip(t *testing.T, entries map[string]io.Reader, symlinkName string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "t.zip")

	f, err := os.Create(p)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	zw := zip.NewWriter(f)
	for name, r := range entries {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.Copy(w, r); err != nil {
			t.Fatal(err)
		}
	}
	if symlinkName != "" {
		h := &zip.FileHeader{Name: symlinkName, Method: zip.Store}
		h.SetMode(os.ModeSymlink | 0o777)
		w, err := zw.CreateHeader(h)
		if err != nil {
			t.Fatal(err)
		}
		_, _ = w.Write([]byte("target"))
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return p
}
