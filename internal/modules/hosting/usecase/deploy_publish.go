package usecase

import (
	"context"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

func (d *Deployer) publishDir(ctx context.Context, root string, prefix string, cacheControl string) error {
	return filepath.WalkDir(root, func(path string, de os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if de.IsDir() {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(rel, "..") {
			return nil
		}
		rel = filepath.ToSlash(rel)
		key := strings.TrimSuffix(prefix, "/") + "/" + strings.TrimPrefix(rel, "/")

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		ct := mime.TypeByExtension(filepath.Ext(path))
		if ct == "" {
			ct = "application/octet-stream"
		}
		return d.obj.Put(ctx, key, f, ct, cacheControl)
	})
}
