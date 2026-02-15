package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"example.com/your-api/internal/modules/hosting/domain"
)

func (s *UploadStore) Get(ctx context.Context, id domain.UploadID) (domain.Upload, error) {
	out, err := s.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(s.table),
		Key:            map[string]types.AttributeValue{"pk": avS(uplPK(id.String()))},
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return domain.Upload{}, err
	}
	if out.Item == nil {
		return domain.Upload{}, domain.ErrUploadNotFound
	}
	return decodeUpload(out.Item), nil
}
