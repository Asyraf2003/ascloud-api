package ports

import (
	"context"
	"time"

	"example.com/your-api/internal/modules/hosting/domain"
)

type SiteStore interface {
	Get(ctx context.Context, id domain.SiteID) (domain.Site, error)
}

type UploadStore interface {
	Put(ctx context.Context, u domain.Upload) error
	Get(ctx context.Context, id domain.UploadID) (domain.Upload, error)
	UpdateStatus(ctx context.Context, id domain.UploadID, status domain.UploadStatus) error
	UpdateStatusAndSize(ctx context.Context, id domain.UploadID, status domain.UploadStatus, sizeBytes int64) error
}

type ObjectStore interface {
	PresignPutZip(ctx context.Context, objectKey string, maxBytes int64, expires time.Duration) (url string, err error)
	Head(ctx context.Context, objectKey string) (sizeBytes int64, err error)
}

type DeployMessage struct {
	SiteID       domain.SiteID   `json:"site_id"`
	UploadID     domain.UploadID `json:"upload_id"`
	ObjectKey    string          `json:"object_key"`
	SizeBytes    int64           `json:"size_bytes"`
	QueuedAtUnix int64           `json:"queued_at_unix"`
}

type DeployQueue interface {
	EnqueueDeploy(ctx context.Context, msg DeployMessage) error
}
