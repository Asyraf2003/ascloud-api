package zipsec

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestExtract_ZipSlipRejected(t *testing.T) {
	z := writeZip(t, map[string]io.Reader{
		"../evil.txt": strings.NewReader("x"),
	}, "")
	_, err := Extract(context.Background(), z, t.TempDir(), DefaultOptions())
	if !errors.Is(err, ErrZipSlip) {
		t.Fatalf("want ErrZipSlip, got %v", err)
	}
}

func TestExtract_SymlinkRejected(t *testing.T) {
	z := writeZip(t, map[string]io.Reader{
		"ok.txt": strings.NewReader("x"),
	}, "link")
	_, err := Extract(context.Background(), z, t.TempDir(), DefaultOptions())
	if !errors.Is(err, ErrZipSymlink) {
		t.Fatalf("want ErrZipSymlink, got %v", err)
	}
}
