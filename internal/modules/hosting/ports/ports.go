package ports

import (
	"context"
	"io"
	"time"

	"example.com/your-api/internal/modules/hosting/domain"
)

type SiteStore interface {
	Put(ctx context.Context, s domain.Site) error
	Get(ctx context.Context, id domain.SiteID) (domain.Site, error)
	List(ctx context.Context, limit int) ([]domain.Site, error)

	UpdateCurrentRelease(ctx context.Context, id domain.SiteID, releaseID domain.ReleaseID) error
}

type UploadStore interface {
	Put(ctx context.Context, u domain.Upload) error
	Get(ctx context.Context, id domain.UploadID) (domain.Upload, error)
	UpdateStatus(ctx context.Context, id domain.UploadID, status domain.UploadStatus) error
	UpdateStatusSizeAndReleaseID(ctx context.Context, id domain.UploadID, status domain.UploadStatus, sizeBytes int64, releaseID domain.ReleaseID) error
}

type ReleaseStore interface {
	Put(ctx context.Context, r domain.Release) error
	Get(ctx context.Context, id domain.ReleaseID) (domain.Release, error)
	ListBySite(ctx context.Context, siteID domain.SiteID, limit int) ([]domain.Release, error)
	UpdateStatus(ctx context.Context, id domain.ReleaseID, status domain.ReleaseStatus, sizeBytes int64, errorCode string) error
}

type EdgeStore interface {
	PutHostMapping(ctx context.Context, host string, siteID domain.SiteID, currentReleaseID domain.ReleaseID, suspended bool) error
}

type ObjectStore interface {
	PresignPutZip(ctx context.Context, objectKey string, maxBytes int64, expires time.Duration) (url string, err error)
	Head(ctx context.Context, objectKey string) (sizeBytes int64, err error)

	Get(ctx context.Context, objectKey string) (body io.ReadCloser, err error)
	Put(ctx context.Context, objectKey string, body io.Reader, contentType string, cacheControl string) error
	Delete(ctx context.Context, objectKey string) error
}

type DeployMessage struct {
	SiteID       domain.SiteID    `json:"site_id"`
	UploadID     domain.UploadID  `json:"upload_id"`
	ReleaseID    domain.ReleaseID `json:"release_id"`
	ObjectKey    string           `json:"object_key"`
	SizeBytes    int64            `json:"size_bytes"`
	QueuedAtUnix int64            `json:"queued_at_unix"`
}

type DeployQueue interface {
	EnqueueDeploy(ctx context.Context, msg DeployMessage) error
}
