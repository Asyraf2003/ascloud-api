package usecase

import (
	"context"
	"fmt"
	"time"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
	"example.com/your-api/internal/shared/apperr"
)

func (d *Deployer) Deploy(ctx context.Context, msg ports.DeployMessage) error {
	if msg.SiteID == "" || msg.UploadID == "" || msg.ReleaseID == "" || msg.ObjectKey == "" {
		return apperr.New("hosting.bad_message", 0, "")
	}

	ctx2, cancel := context.WithTimeout(ctx, d.cfg.DeployTimeout)
	defer cancel()

	d.initRelease(ctx2, msg)

	u, err := d.up.Get(ctx2, msg.UploadID)
	if err != nil {
		return apperr.Wrap(err, "hosting.upload_not_found", 0, "")
	}
	if u.SiteID != msg.SiteID {
		return d.permanentFail(ctx2, msg, "hosting.site_mismatch")
	}
	if u.Status == domain.UploadStatusDeployed {
		return nil
	}
	if u.Status != domain.UploadStatusQueued {
		return d.permanentFail(ctx2, msg, "hosting.upload_not_queued")
	}
	if d.cfg.MaxZipBytes > 0 && msg.SizeBytes > d.cfg.MaxZipBytes {
		return d.permanentFail(ctx2, msg, "hosting.zip_too_large")
	}

	zipPath, err := d.downloadZip(ctx2, msg.ObjectKey, msg.UploadID, d.cfg.MaxZipBytes)
	if err != nil {
		if err == ErrZipTooLarge {
			return d.permanentFail(ctx2, msg, "hosting.zip_too_large")
		}
		return apperr.Wrap(err, "hosting.s3_get_failed", 0, "")
	}

	extractDir, res, err := d.extractZip(ctx2, zipPath, msg.ReleaseID)
	if err != nil {
		if code := zipErrCode(err); code != "" {
			return d.permanentFail(ctx2, msg, code)
		}
		return apperr.Wrap(err, "hosting.extract_failed", 0, "")
	}

	prefix := fmt.Sprintf("sites/%s/releases/%s", msg.SiteID.String(), msg.ReleaseID.String())
	if err := d.publishDir(ctx2, extractDir, prefix, d.cfg.CacheControl); err != nil {
		return apperr.Wrap(err, "hosting.s3_put_failed", 0, "")
	}

	nowSize := res.TotalBytes
	if err := d.rel.UpdateStatus(ctx2, msg.ReleaseID, domain.ReleaseStatusSuccess, nowSize, ""); err != nil {
		return apperr.Wrap(err, "hosting.ddb_release_update_failed", 0, "")
	}
	if err := d.sites.UpdateCurrentRelease(ctx2, msg.SiteID, msg.ReleaseID); err != nil {
		return apperr.Wrap(err, "hosting.ddb_site_update_failed", 0, "")
	}
	if err := d.up.UpdateStatusSizeAndReleaseID(ctx2, msg.UploadID, domain.UploadStatusDeployed, nowSize, msg.ReleaseID); err != nil {
		return apperr.Wrap(err, "hosting.ddb_upload_update_failed", 0, "")
	}

	_ = d.obj.Delete(ctx2, msg.ObjectKey)
	_ = time.Now() // keep imports stable under gofmt if you remove time usage later
	return nil
}
