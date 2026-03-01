package dynamodb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
)

type SiteStore struct {
	db    *dynamodb.Client
	table string
}

func NewSiteStore(db *dynamodb.Client, table string) *SiteStore {
	return &SiteStore{db: db, table: table}
}

func (s *SiteStore) Get(ctx context.Context, id domain.SiteID) (domain.Site, error) {
	out, err := s.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      aws.String(s.table),
		Key:            map[string]types.AttributeValue{"pk": avS(sitePK(id.String()))},
		ConsistentRead: aws.Bool(true),
	})
	if err != nil {
		return domain.Site{}, err
	}
	if out.Item == nil {
		return domain.Site{}, domain.ErrSiteNotFound
	}
	return decodeSite(out.Item), nil
}

func (s *SiteStore) UpdateCurrentRelease(ctx context.Context, id domain.SiteID, releaseID domain.ReleaseID) error {
	now := time.Now().UTC().Unix()
	_, err := s.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.table),
		Key:       map[string]types.AttributeValue{"pk": avS(sitePK(id.String()))},
		UpdateExpression: aws.String(
			"SET site_id = if_not_exists(site_id, :sid), " +
				"current_release_id = :rid, " +
				"suspended = if_not_exists(suspended, :z), " +
				"created_at = if_not_exists(created_at, :t), " +
				"updated_at = :t",
		),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":sid": avS(id.String()),
			":rid": avS(releaseID.String()),
			":z":   avN(int64(0)),
			":t":   avN(now),
		},
	})
	return err
}

func decodeSite(m map[string]types.AttributeValue) domain.Site {
	suspended := getN(m, "suspended") > 0
	return domain.Site{
		ID:             domain.SiteID(getS(m, "site_id")),
		CurrentRelease: domain.ReleaseID(getS(m, "current_release_id")),
		Suspended:      suspended,
		CreatedAt:      time.Unix(getN(m, "created_at"), 0).UTC(),
	}
}

var _ ports.SiteStore = (*SiteStore)(nil)
