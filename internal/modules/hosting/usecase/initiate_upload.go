package usecase

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"example.com/your-api/internal/modules/hosting/domain"
)

const requiredZipContentType = "application/zip"

type InitiateUploadOutput struct {
	UploadID      domain.UploadID
	ObjectKey     string
	PutURL        string
	ExpiresAtUnix int64
	MaxBytes      int64
	ContentType   string
}

func (s *Service) InitiateUpload(ctx context.Context, siteID domain.SiteID) (InitiateUploadOutput, error) {
	sid := strings.TrimSpace(siteID.String())
	uid := domain.UploadID(uuid.NewString())
	key := fmt.Sprintf("sites/%s/uploads/%s.zip", sid, uid.String())

	url, err := s.obj.PresignPutZip(ctx, key, s.cfg.MaxUploadBytes, s.cfg.PresignTTL)
	if err != nil {
		return InitiateUploadOutput{}, err
	}

	now := time.Now().UTC()
	err = s.up.Put(ctx, domain.Upload{
		ID:        uid,
		SiteID:    domain.SiteID(sid),
		ObjectKey: key,
		SizeBytes: 0,
		Status:    domain.UploadStatusInitiated,
		CreatedAt: now,
	})
	if err != nil {
		return InitiateUploadOutput{}, err
	}

	return InitiateUploadOutput{
		UploadID:      uid,
		ObjectKey:     key,
		PutURL:        url,
		ExpiresAtUnix: now.Add(s.cfg.PresignTTL).Unix(),
		MaxBytes:      s.cfg.MaxUploadBytes,
		ContentType:   requiredZipContentType,
	}, nil
}
