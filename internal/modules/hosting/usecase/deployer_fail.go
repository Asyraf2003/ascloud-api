package usecase

import (
	"context"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
	"example.com/your-api/internal/shared/apperr"
)

func (d *Deployer) permanentFail(ctx context.Context, msg ports.DeployMessage, code string) error {
	_ = d.up.UpdateStatus(ctx, msg.UploadID, domain.UploadStatusFailed)
	_ = d.rel.UpdateStatus(ctx, msg.ReleaseID, domain.ReleaseStatusFailed, 0, code)
	return apperr.New(code, 0, "")
}
