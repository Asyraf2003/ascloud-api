package dynamodb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"example.com/your-api/internal/modules/hosting/domain"
)

func (s *UploadStore) Put(ctx context.Context, u domain.Upload) error {
	now := time.Now().UTC()
	if u.CreatedAt.IsZero() {
		u.CreatedAt = now
	}
	item := map[string]types.AttributeValue{
		"pk":         avS(uplPK(u.ID.String())),
		"upload_id":  avS(u.ID.String()),
		"site_id":    avS(u.SiteID.String()),
		"object_key": avS(u.ObjectKey),
		"size_bytes": avN(u.SizeBytes),
		"status":     avS(string(u.Status)),
		"created_at": avN(u.CreatedAt.UTC().Unix()),
	}
	_, err := s.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(s.table),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(pk)"),
	})
	return err
}
