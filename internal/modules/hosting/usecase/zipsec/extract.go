package zipsec

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"strings"
)

var disallowedExt = map[string]struct{}{
	".exe":   {},
	".dll":   {},
	".so":    {},
	".dylib": {},
	".bat":   {},
	".cmd":   {},
	".ps1":   {},
	".sh":    {},
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

	var (
		res       Result
		violSet   = map[error]struct{}{}
		seenFiles int
	)

	addViolation := func(v error) {
		if v == nil {
			return
		}
		violSet[v] = struct{}{}
	}

	for _, f := range zr.File {
		if err := ctx.Err(); err != nil {
			return res, err
		}

		clean, cerr := cleanZipPath(f.Name, opt.MaxDepth)
		if cerr != nil {
			addViolation(cerr)
			continue
		}
		if clean == "" {
			continue
		}

		if isSymlinkOrSpecial(f) {
			addViolation(ErrZipSymlink)
			continue
		}

		full := filepath.Join(dest, clean)
		rel, rerr := filepath.Rel(dest, full)
		if rerr != nil || strings.HasPrefix(rel, "..") {
			addViolation(ErrZipSlip)
			continue
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(full, 0o755); err != nil {
				return res, err
			}
			continue
		}

		// count entries regardless extracted/blocked (anti zip spam)
		seenFiles++
		if opt.MaxFiles > 0 && seenFiles > opt.MaxFiles {
			addViolation(ErrTooManyFiles)
			break
		}

		// denylist minimal executable/script
		if isDisallowedByExt(clean) {
			addViolation(ErrDisallowedFile)
			continue
		}

		allow, aerr := allowedBytes(opt, res.TotalBytes, int64(f.UncompressedSize64))
		if aerr != nil {
			// quota is terminal (avoid heavy work)
			addViolation(aerr)
			break
		}

		n, xerr := extractOne(ctx, f, full, allow)
		if xerr != nil {
			// I/O failure is not a "violation": surface as error (transient)
			return res, xerr
		}
		res.Files++
		res.TotalBytes += n
	}

	if opt.MaxTotalBytes > 0 && res.TotalBytes > opt.MaxTotalBytes {
		addViolation(ErrOverQuota)
	}

	if len(violSet) > 0 {
		// keep legacy behavior for single-violation tests:
		// return the sentinel directly if only 1 unique violation.
		if len(violSet) == 1 {
			for v := range violSet {
				return res, v
			}
		}

		errs := make([]error, 0, len(violSet))
		for v := range violSet {
			errs = append(errs, v)
		}
		return res, &ViolationsError{Errs: errs}
	}

	return res, nil
}

func isDisallowedByExt(clean string) bool {
	ext := strings.ToLower(filepath.Ext(clean))
	if ext == "" {
		return false
	}
	_, bad := disallowedExt[ext]
	return bad
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

// TODO justify: >100 lines sementara; akan dipecah saat milestone 9.
