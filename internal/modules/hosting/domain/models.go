package domain

import "time"

type Site struct {
	ID             SiteID
	CreatedAt      time.Time
	Suspended      bool
	CurrentRelease ReleaseID
	Host           string
	UpdatedAt      time.Time
}

type UploadStatus string

const (
	UploadStatusInitiated UploadStatus = "initiated"
	UploadStatusUploaded  UploadStatus = "uploaded"
	UploadStatusQueued    UploadStatus = "queued"
	UploadStatusDeployed  UploadStatus = "deployed"
	UploadStatusFailed    UploadStatus = "failed"
)

type Upload struct {
	ID        UploadID
	SiteID    SiteID
	ReleaseID ReleaseID // empty means not assigned yet
	ObjectKey string
	SizeBytes int64
	Status    UploadStatus
	CreatedAt time.Time
}
