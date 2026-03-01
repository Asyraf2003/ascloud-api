package usecase

import (
	"context"
	"os"
	"path/filepath"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/usecase/zipsec"
)

func (d *Deployer) extractZip(ctx context.Context, zipPath string, releaseID domain.ReleaseID) (string, zipsec.Result, error) {
	dir := filepath.Join(d.cfg.TmpDir, "release-"+releaseID.String())
	_ = os.RemoveAll(dir)
	res, err := zipsec.Extract(ctx, zipPath, dir, d.cfg.Extract)
	return dir, res, err
}
