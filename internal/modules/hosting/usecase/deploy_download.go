package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"example.com/your-api/internal/modules/hosting/domain"
)

var ErrZipTooLarge = errors.New("zip too large")

func (d *Deployer) downloadZip(ctx context.Context, objectKey string, uploadID domain.UploadID, max int64) (string, error) {
	rc, err := d.obj.Get(ctx, objectKey)
	if err != nil {
		return "", err
	}
	defer rc.Close()

	p := filepath.Join(d.cfg.TmpDir, fmt.Sprintf("upload-%s.zip", uploadID.String()))
	f, err := os.OpenFile(p, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := copyLimited(ctx, f, rc, max); err != nil {
		_ = os.Remove(p)
		return "", err
	}
	return p, nil
}

func copyLimited(ctx context.Context, dst io.Writer, src io.Reader, max int64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if max <= 0 {
		_, err := io.Copy(dst, src)
		return err
	}

	lr := io.LimitReader(src, max+1)
	n, err := io.Copy(dst, lr)
	if err != nil {
		return err
	}
	if n > max {
		return ErrZipTooLarge
	}
	return nil
}
