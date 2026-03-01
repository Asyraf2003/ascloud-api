package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
)

type CompleteUploadOutput struct {
	Status domain.UploadStatus
	Size   int64
}

func (s *Service) CompleteUpload(ctx context.Context, siteID domain.SiteID, uploadID domain.UploadID) (CompleteUploadOutput, error) {
	u, err := s.up.Get(ctx, uploadID)
	if err != nil {
		return CompleteUploadOutput{}, err
	}
	if u.SiteID != siteID {
		return CompleteUploadOutput{}, domain.ErrUploadSiteMismatch
	}
	if u.Status == domain.UploadStatusQueued {
		return CompleteUploadOutput{Status: u.Status, Size: u.SizeBytes}, nil
	}

	size, err := s.obj.Head(ctx, u.ObjectKey)
	if err != nil {
		_ = s.up.UpdateStatus(ctx, uploadID, domain.UploadStatusFailed)
		return CompleteUploadOutput{}, err
	}
	if s.cfg.MaxUploadBytes > 0 && size > s.cfg.MaxUploadBytes {
		_ = s.up.UpdateStatusSizeAndReleaseID(ctx, uploadID, domain.UploadStatusFailed, size, "")
		return CompleteUploadOutput{}, domain.ErrUploadTooLarge
	}

	rid := domain.ReleaseID(uuid.NewString())

	if err := s.up.UpdateStatusSizeAndReleaseID(ctx, uploadID, domain.UploadStatusQueued, size, rid); err != nil {
		return CompleteUploadOutput{}, err
	}

	msg := ports.DeployMessage{
		SiteID:       siteID,
		UploadID:     uploadID,
		ReleaseID:    rid,
		ObjectKey:    u.ObjectKey,
		SizeBytes:    size,
		QueuedAtUnix: time.Now().UTC().Unix(),
	}
	if err := s.queue.EnqueueDeploy(ctx, msg); err != nil {
		_ = s.up.UpdateStatus(ctx, uploadID, domain.UploadStatusFailed)
		return CompleteUploadOutput{}, err
	}

	return CompleteUploadOutput{Status: domain.UploadStatusQueued, Size: size}, nil
}
