package domain

import "time"

type Site struct {
	ID             SiteID
	CreatedAt      time.Time
	Suspended      bool
	CurrentRelease ReleaseID // empty means none yet
}

type UploadStatus string

const (
	UploadStatusInitiated UploadStatus = "initiated"
	UploadStatusUploaded  UploadStatus = "uploaded"
	UploadStatusQueued    UploadStatus = "queued"
	UploadStatusFailed    UploadStatus = "failed"
)

type Upload struct {
	ID        UploadID
	SiteID    SiteID
	ObjectKey string
	SizeBytes int64
	Status    UploadStatus
	CreatedAt time.Time
}
