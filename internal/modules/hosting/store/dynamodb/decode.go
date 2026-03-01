package dynamodb

import (
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"example.com/your-api/internal/modules/hosting/domain"
)

func decodeUpload(m map[string]types.AttributeValue) domain.Upload {
	return domain.Upload{
		ID:        domain.UploadID(getS(m, "upload_id")),
		SiteID:    domain.SiteID(getS(m, "site_id")),
		ReleaseID: domain.ReleaseID(getS(m, "release_id")),
		ObjectKey: getS(m, "object_key"),
		SizeBytes: getN(m, "size_bytes"),
		Status:    domain.UploadStatus(getS(m, "status")),
		CreatedAt: time.Unix(getN(m, "created_at"), 0).UTC(),
	}
}
