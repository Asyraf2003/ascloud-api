package usecase

import (
	"context"
	"fmt"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
	"example.com/your-api/internal/shared/apperr"
)

func (d *Deployer) Deploy(ctx context.Context, msg ports.DeployMessage) error {
	if err := validateDeployMessage(msg); err != nil {
		return err
	}

	ctx2, cancel := context.WithTimeout(ctx, d.cfg.DeployTimeout)
	defer cancel()

	d.initRelease(ctx2, msg)

	u, err := d.loadUploadForDeploy(ctx2, msg)
	if err != nil {
		return err
	}
	if u.Status == domain.UploadStatusDeployed {
		return nil
	}

	size := msg.SizeBytes
	if size <= 0 {
		size = u.SizeBytes
	}
	if d.cfg.MaxZipBytes > 0 && size > d.cfg.MaxZipBytes {
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

	if err := d.markSuccess(ctx2, msg, res.TotalBytes); err != nil {
		return err
	}

	_ = d.obj.Delete(ctx2, msg.ObjectKey)
	return nil
}

func validateDeployMessage(msg ports.DeployMessage) error {
	if msg.SiteID == "" || msg.UploadID == "" || msg.ReleaseID == "" || msg.ObjectKey == "" {
		return apperr.New("hosting.bad_message", 0, "")
	}
	return nil
}
