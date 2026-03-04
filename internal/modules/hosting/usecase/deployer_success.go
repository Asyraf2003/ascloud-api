package usecase

import (
	"context"
	"time"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
	"example.com/your-api/internal/shared/apperr"
)

func (d *Deployer) markSuccess(ctx context.Context, msg ports.DeployMessage, sizeBytes int64) error {
	if err := d.rel.UpdateStatus(ctx, msg.ReleaseID, domain.ReleaseStatusSuccess, sizeBytes, "", nil); err != nil {
		return apperr.Wrap(err, "hosting.ddb_release_update_failed", 0, "")
	}
	if err := d.sites.UpdateCurrentRelease(ctx, msg.SiteID, msg.ReleaseID); err != nil {
		return apperr.Wrap(err, "hosting.ddb_site_update_failed", 0, "")
	}
	if err := d.up.UpdateStatusSizeAndReleaseID(ctx, msg.UploadID, domain.UploadStatusDeployed, sizeBytes, msg.ReleaseID); err != nil {
		return apperr.Wrap(err, "hosting.ddb_upload_update_failed", 0, "")
	}

	if d.audit != nil {
		_ = d.audit.Record(ctx, ports.AuditEvent{
			SiteID: msg.SiteID,
			Event:  "hosting_deploy_succeeded",
			At:     time.Now(),
			Meta: map[string]any{
				"request_id": msg.RequestID,
				"upload_id":  msg.UploadID,
				"release_id": msg.ReleaseID,
				"size_bytes": sizeBytes,
			},
		})
	}

	return nil
}
