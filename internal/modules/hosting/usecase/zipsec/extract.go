package zipsec

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var (
	ErrZipSlip      = errors.New("zip slip")
	ErrZipSymlink   = errors.New("zip symlink")
	ErrTooManyFiles = errors.New("zip too many files")
	ErrTooDeep      = errors.New("zip too deep")
	ErrOverQuota    = errors.New("zip over quota")
)

type Options struct {
	MaxTotalBytes int64
	MaxFiles      int
	MaxDepth      int
	MaxFileBytes  int64
}

type Result struct {
	Files      int
	TotalBytes int64
}

func DefaultOptions() Options {
	return Options{
		MaxTotalBytes: 20 * 1024 * 1024,
		MaxFiles:      2000,
		MaxDepth:      20,
		MaxFileBytes:  20 * 1024 * 1024,
	}
}

func Extract(ctx context.Context, zipPath, dest string, opt Options) (Result, error) {
	zr, err := zip.OpenReader(zipPath)
	if err != nil {
		return Result{}, err
	}
	defer zr.Close()

	if err := os.MkdirAll(dest, 0o755); err != nil {
		return Result{}, err
	}

	var res Result
	for _, f := range zr.File {
		if err := ctx.Err(); err != nil {
			return res, err
		}
		clean, err := cleanZipPath(f.Name, opt.MaxDepth)
		if err != nil {
			return res, err
		}
		if clean == "" {
			continue
		}
		if isSymlinkOrSpecial(f) {
			return res, ErrZipSymlink
		}

		full := filepath.Join(dest, clean)
		if rel, _ := filepath.Rel(dest, full); strings.HasPrefix(rel, "..") {
			return res, ErrZipSlip
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(full, 0o755); err != nil {
				return res, err
			}
			continue
		}

		if opt.MaxFiles > 0 && res.Files+1 > opt.MaxFiles {
			return res, ErrTooManyFiles
		}

		allow := opt.MaxFileBytes
		if opt.MaxTotalBytes > 0 {
			rem := opt.MaxTotalBytes - res.TotalBytes
			if allow == 0 || rem < allow {
				allow = rem
			}
		}
		if allow > 0 && int64(f.UncompressedSize64) > allow {
			return res, ErrOverQuota
		}

		n, err := extractOne(ctx, f, full, allow)
		if err != nil {
			return res, err
		}
		res.Files++
		res.TotalBytes += n
	}
	if opt.MaxTotalBytes > 0 && res.TotalBytes > opt.MaxTotalBytes {
		return res, ErrOverQuota
	}
	return res, nil
}

func cleanZipPath(name string, maxDepth int) (string, error) {
	if strings.Contains(name, `\`) {
		return "", ErrZipSlip
	}
	n := strings.TrimPrefix(name, "/")
	c := filepath.Clean(n)
	if c == "." {
		return "", nil
	}
	if filepath.IsAbs(c) || strings.HasPrefix(c, "..") {
		return "", ErrZipSlip
	}
	if maxDepth > 0 && strings.Count(c, string(os.PathSeparator)) > maxDepth {
		return "", ErrTooDeep
	}
	return c, nil
}

func isSymlinkOrSpecial(f *zip.File) bool {
	m := f.Mode()
	if m&os.ModeSymlink != 0 {
		return true
	}
	return m&os.ModeType != 0 && !f.FileInfo().IsDir()
}

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
