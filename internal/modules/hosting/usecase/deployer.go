package usecase

import (
	"context"
	"fmt"
	"time"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
	"example.com/your-api/internal/modules/hosting/usecase/zipsec"
	"example.com/your-api/internal/shared/apperr"
)

type Deployer struct {
	cfg   DeployConfig
	up    ports.UploadStore
	sites ports.SiteStore
	rel   ports.ReleaseStore
	obj   ports.ObjectStore
}

func NewDeployer(cfg DeployConfig, up ports.UploadStore, sites ports.SiteStore, rel ports.ReleaseStore, obj ports.ObjectStore) *Deployer {
	return &Deployer{cfg: cfg, up: up, sites: sites, rel: rel, obj: obj}
}

func (d *Deployer) Deploy(ctx context.Context, msg ports.DeployMessage) error {
	if msg.SiteID == "" || msg.UploadID == "" || msg.ReleaseID == "" || msg.ObjectKey == "" {
		return apperr.New("hosting.bad_message", 0, "")
	}

	ctx2, cancel := context.WithTimeout(ctx, d.cfg.DeployTimeout)
	defer cancel()

	_ = d.rel.Put(ctx2, domain.Release{
		ID:        msg.ReleaseID,
		SiteID:    msg.SiteID,
		Status:    domain.ReleaseStatusPending,
		SizeBytes: 0,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	})

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

	if err := d.rel.UpdateStatus(ctx2, msg.ReleaseID, domain.ReleaseStatusSuccess, res.TotalBytes, ""); err != nil {
		return apperr.Wrap(err, "hosting.ddb_release_update_failed", 0, "")
	}
	if err := d.sites.UpdateCurrentRelease(ctx2, msg.SiteID, msg.ReleaseID); err != nil {
		return apperr.Wrap(err, "hosting.ddb_site_update_failed", 0, "")
	}
	if err := d.up.UpdateStatusSizeAndReleaseID(ctx2, msg.UploadID, domain.UploadStatusDeployed, res.TotalBytes, msg.ReleaseID); err != nil {
		return apperr.Wrap(err, "hosting.ddb_upload_update_failed", 0, "")
	}

	_ = d.obj.Delete(ctx2, msg.ObjectKey)
	return nil
}

func (d *Deployer) permanentFail(ctx context.Context, msg ports.DeployMessage, code string) error {
	_ = d.up.UpdateStatus(ctx, msg.UploadID, domain.UploadStatusFailed)
	_ = d.rel.UpdateStatus(ctx, msg.ReleaseID, domain.ReleaseStatusFailed, 0, code)
	return apperr.New(code, 0, "")
}

func zipErrCode(err error) string {
	switch err {
	case zipsec.ErrZipSlip:
		return "hosting.zip_slip"
	case zipsec.ErrZipSymlink:
		return "hosting.zip_symlink"
	case zipsec.ErrTooManyFiles:
		return "hosting.zip_too_many_files"
	case zipsec.ErrTooDeep:
		return "hosting.zip_too_deep"
	case zipsec.ErrOverQuota:
		return "hosting.extract_over_quota"
	default:
		return ""
	}
}
