package zipsec

import (
	"context"
	"errors"
	"io"
	"testing"
)

func TestExtract_OverQuota(t *testing.T) {
	z := writeZip(t, map[string]io.Reader{
		"big.bin": io.LimitReader(zeroReader{}, 2048),
	}, "")
	opt := DefaultOptions()
	opt.MaxTotalBytes = 1024
	opt.MaxFileBytes = 1024
	_, err := Extract(context.Background(), z, t.TempDir(), opt)
	if !errors.Is(err, ErrOverQuota) {
		t.Fatalf("want ErrOverQuota, got %v", err)
	}
}
