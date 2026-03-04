package usecase

import (
	"context"
	"time"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
	"example.com/your-api/internal/shared/apperr"
)

func (d *Deployer) permanentFail(ctx context.Context, msg ports.DeployMessage, code string, violations []string) error {
	_ = d.up.UpdateStatus(ctx, msg.UploadID, domain.UploadStatusFailed)
	_ = d.rel.UpdateStatus(ctx, msg.ReleaseID, domain.ReleaseStatusFailed, 0, code, violations)

	if d.audit != nil {
		meta := map[string]any{
			"request_id": msg.RequestID,
			"upload_id":  msg.UploadID,
			"release_id": msg.ReleaseID,
			"error_code": code,
		}
		if len(violations) > 0 {
			meta["violations"] = violations
		}

		_ = d.audit.Record(ctx, ports.AuditEvent{
			SiteID: msg.SiteID,
			Event:  "hosting_deploy_failed",
			At:     time.Now(),
			Meta:   meta,
		})
	}

	return apperr.New(code, 0, "")
}
