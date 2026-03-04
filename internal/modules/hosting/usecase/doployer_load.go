package usecase

import (
	"context"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
	"example.com/your-api/internal/shared/apperr"
)

func (d *Deployer) loadUploadForDeploy(ctx context.Context, msg ports.DeployMessage) (domain.Upload, error) {
	u, err := d.up.Get(ctx, msg.UploadID)
	if err != nil {
		return domain.Upload{}, apperr.Wrap(err, "hosting.upload_not_found", 0, "")
	}
	if u.SiteID != msg.SiteID {
		return domain.Upload{}, d.permanentFail(ctx, msg, "hosting.site_mismatch", nil)
	}
	if u.Status == domain.UploadStatusDeployed {
		return u, nil
	}
	if u.Status != domain.UploadStatusQueued {
		return domain.Upload{}, d.permanentFail(ctx, msg, "hosting.upload_not_queued", nil)
	}
	return u, nil
}
