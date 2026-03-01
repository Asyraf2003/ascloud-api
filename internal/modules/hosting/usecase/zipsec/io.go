package zipsec

import (
	"archive/zip"
	"context"
	"io"
	"os"
	"path/filepath"
)

func extractOne(ctx context.Context, f *zip.File, full string, allow int64) (int64, error) {
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		return 0, err
	}

	rc, err := f.Open()
	if err != nil {
		return 0, err
	}
	defer rc.Close()

	out, err := os.OpenFile(full, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return 0, err
	}
	defer out.Close()

	n, err := copyAtMost(ctx, out, rc, allow)
	if err != nil {
		_ = os.Remove(full)
		return n, err
	}
	return n, nil
}

func copyAtMost(ctx context.Context, dst io.Writer, src io.Reader, allow int64) (int64, error) {
	max := int64(0)
	if allow > 0 {
		max = allow + 1
	}

	buf := make([]byte, 32*1024)
	var n int64
	for {
		if err := ctx.Err(); err != nil {
			return n, err
		}
		if max > 0 && n >= max {
			break
		}

		toRead := len(buf)
		if max > 0 {
			rem := max - n
			if rem <= 0 {
				break
			}
			if rem < int64(toRead) {
				toRead = int(rem)
			}
		}

		m, rerr := src.Read(buf[:toRead])
		if m > 0 {
			w, werr := dst.Write(buf[:m])
			n += int64(w)
			if werr != nil {
				return n, werr
			}
			if w != m {
				return n, io.ErrShortWrite
			}
		}
		if rerr == io.EOF {
			break
		}
		if rerr != nil {
			return n, rerr
		}
	}

	if allow > 0 && n > allow {
		return n, ErrOverQuota
	}
	return n, nil
}
