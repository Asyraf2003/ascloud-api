package dynamodb

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"example.com/your-api/internal/modules/hosting/domain"
)

func (s *UploadStore) UpdateStatus(ctx context.Context, id domain.UploadID, status domain.UploadStatus) error {
	_, err := s.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:           aws.String(s.table),
		Key:                 map[string]types.AttributeValue{"pk": avS(uplPK(id.String()))},
		ConditionExpression: aws.String("attribute_exists(pk)"),
		UpdateExpression:    aws.String("SET #st = :s, updated_at = :t"),
		ExpressionAttributeNames: map[string]string{
			"#st": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":s": avS(string(status)),
			":t": avN(time.Now().UTC().Unix()),
		},
	})
	if err == nil {
		return nil
	}
	var cfe *types.ConditionalCheckFailedException
	if errors.As(err, &cfe) {
		return domain.ErrUploadNotFound
	}
	return err
}

func (s *UploadStore) UpdateStatusSizeAndReleaseID(ctx context.Context, id domain.UploadID, status domain.UploadStatus, sizeBytes int64, releaseID domain.ReleaseID) error {
	_, err := s.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:           aws.String(s.table),
		Key:                 map[string]types.AttributeValue{"pk": avS(uplPK(id.String()))},
		ConditionExpression: aws.String("attribute_exists(pk)"),
		UpdateExpression:    aws.String("SET #st = :s, size_bytes = :b, release_id = :r, updated_at = :t"),
		ExpressionAttributeNames: map[string]string{
			"#st": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":s": avS(string(status)),
			":b": avN(sizeBytes),
			":r": avS(releaseID.String()),
			":t": avN(time.Now().UTC().Unix()),
		},
	})
	if err == nil {
		return nil
	}
	var cfe *types.ConditionalCheckFailedException
	if errors.As(err, &cfe) {
		return domain.ErrUploadNotFound
	}
	return err
}
