package zipsec

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
)

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
		rel, rerr := filepath.Rel(dest, full)
		if rerr != nil || strings.HasPrefix(rel, "..") {
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

		allow, err := allowedBytes(opt, res.TotalBytes, int64(f.UncompressedSize64))
		if err != nil {
			return res, err
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

func allowedBytes(opt Options, used int64, fileSize int64) (int64, error) {
	allow := opt.MaxFileBytes

	if opt.MaxTotalBytes > 0 {
		rem := opt.MaxTotalBytes - used
		if rem < 0 {
			return 0, ErrOverQuota
		}
		if allow == 0 || rem < allow {
			allow = rem
		}
	}

	if allow > 0 && fileSize > allow {
		return 0, ErrOverQuota
	}
	return allow, nil
}
