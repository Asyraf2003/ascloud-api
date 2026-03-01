package dynamodb

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
)

type ReleaseStore struct {
	db    *dynamodb.Client
	table string
}

func NewReleaseStore(db *dynamodb.Client, table string) *ReleaseStore {
	return &ReleaseStore{db: db, table: table}
}

func (s *ReleaseStore) Put(ctx context.Context, r domain.Release) error {
	now := time.Now().UTC()
	if r.CreatedAt.IsZero() {
		r.CreatedAt = now
	}
	if r.UpdatedAt.IsZero() {
		r.UpdatedAt = r.CreatedAt
	}
	item := map[string]types.AttributeValue{
		"pk":         avS(relPK(r.ID.String())),
		"release_id": avS(r.ID.String()),
		"site_id":    avS(r.SiteID.String()),
		"status":     avS(string(r.Status)),
		"size_bytes": avN(r.SizeBytes),
		"error_code": avS(r.ErrorCode),
		"created_at": avN(r.CreatedAt.UTC().Unix()),
		"updated_at": avN(r.UpdatedAt.UTC().Unix()),
	}
	_, err := s.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(s.table),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(pk)"),
	})
	var cfe *types.ConditionalCheckFailedException
	if errors.As(err, &cfe) {
		return nil
	}
	return err
}

func (s *ReleaseStore) UpdateStatus(ctx context.Context, id domain.ReleaseID, status domain.ReleaseStatus, sizeBytes int64, errorCode string) error {
	_, err := s.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:           aws.String(s.table),
		Key:                 map[string]types.AttributeValue{"pk": avS(relPK(id.String()))},
		ConditionExpression: aws.String("attribute_exists(pk)"),
		UpdateExpression:    aws.String("SET #st = :s, size_bytes = :b, error_code = :e, updated_at = :t"),
		ExpressionAttributeNames: map[string]string{
			"#st": "status",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":s": avS(string(status)),
			":b": avN(sizeBytes),
			":e": avS(errorCode),
			":t": avN(time.Now().UTC().Unix()),
		},
	})
	if err == nil {
		return nil
	}
	var cfe *types.ConditionalCheckFailedException
	if errors.As(err, &cfe) {
		return domain.ErrReleaseNotFound
	}
	return err
}

var _ ports.ReleaseStore = (*ReleaseStore)(nil)
