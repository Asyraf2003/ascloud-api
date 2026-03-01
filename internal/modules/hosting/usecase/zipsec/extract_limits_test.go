package zipsec

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
)

func TestExtract_TooManyFiles(t *testing.T) {
	z := writeZip(t, map[string]io.Reader{
		"a.txt": strings.NewReader("a"),
		"b.txt": strings.NewReader("b"),
	}, "")
	opt := DefaultOptions()
	opt.MaxFiles = 1
	_, err := Extract(context.Background(), z, t.TempDir(), opt)
	if !errors.Is(err, ErrTooManyFiles) {
		t.Fatalf("want ErrTooManyFiles, got %v", err)
	}
}

func TestExtract_TooDeep(t *testing.T) {
	z := writeZip(t, map[string]io.Reader{
		"a/b/c.txt": strings.NewReader("x"),
	}, "")
	opt := DefaultOptions()
	opt.MaxDepth = 1
	_, err := Extract(context.Background(), z, t.TempDir(), opt)
	if !errors.Is(err, ErrTooDeep) {
		t.Fatalf("want ErrTooDeep, got %v", err)
	}
}
