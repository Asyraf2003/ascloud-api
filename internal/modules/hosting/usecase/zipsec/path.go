package zipsec

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
)

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
