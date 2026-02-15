package http

import (
	"context"

	"example.com/your-api/internal/modules/hosting/usecase"
)

type UploadFlow interface {
	InitiateUpload(ctx context.Context, siteID string) (usecase.InitiateUploadOutput, error)
	CompleteUpload(ctx context.Context, siteID string, uploadID string) (usecase.CompleteUploadOutput, error)
}

type InitiateUploadResponse struct {
	UploadID      string `json:"upload_id"`
	ObjectKey     string `json:"object_key"`
	PutURL        string `json:"put_url"`
	ExpiresAtUnix int64  `json:"expires_at_unix"`
	MaxBytes      int64  `json:"max_bytes"`
}

type CompleteUploadResponse struct {
	Status    string `json:"status"`
	SizeBytes int64  `json:"size_bytes"`
}
