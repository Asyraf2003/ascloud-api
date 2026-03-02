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

type SiteStore struct {
	db    *dynamodb.Client
	table string
}

func NewSiteStore(db *dynamodb.Client, table string) *SiteStore {
	return &SiteStore{db: db, table: table}
}

func (s *SiteStore) Put(ctx context.Context, site domain.Site) error {
	now := time.Now().UTC()
	if site.CreatedAt.IsZero() {
		site.CreatedAt = now
	}
	if site.UpdatedAt.IsZero() {
		site.UpdatedAt = site.CreatedAt
	}

	susp := int64(0)
	if site.Suspended {
		susp = 1
	}

	item := map[string]types.AttributeValue{
		"pk":                 avS(sitePK(site.ID.String())),
		"site_id":            avS(site.ID.String()),
		"host":               avS(site.Host),
		"current_release_id": avS(site.CurrentRelease.String()),
		"suspended":          avN(susp),
		"created_at":         avN(site.CreatedAt.UTC().Unix()),
		"updated_at":         avN(site.UpdatedAt.UTC().Unix()),
	}

	_, err := s.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(s.table),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(pk)"),
	})

	var cfe *types.ConditionalCheckFailedException
	if errors.As(err, &cfe) {
		return domain.ErrSiteAlreadyExists
	}
	return err
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

func (s *SiteStore) List(ctx context.Context, limit int) ([]domain.Site, error) {
	in := &dynamodb.ScanInput{
		TableName:                 aws.String(s.table),
		FilterExpression:          aws.String("begins_with(pk, :pfx)"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":pfx": avS("site#")},
	}
	if limit > 0 {
		in.Limit = aws.Int32(int32(limit))
	}

	out, err := s.db.Scan(ctx, in)
	if err != nil {
		return nil, err
	}

	sites := make([]domain.Site, 0, len(out.Items))
	for _, it := range out.Items {
		sites = append(sites, decodeSite(it))
	}
	return sites, nil
}

func (s *SiteStore) UpdateCurrentRelease(ctx context.Context, id domain.SiteID, releaseID domain.ReleaseID) error {
	now := time.Now().UTC().Unix()

	_, err := s.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(s.table),
		Key:       map[string]types.AttributeValue{"pk": avS(sitePK(id.String()))},
		UpdateExpression: aws.String(
			"SET site_id = if_not_exists(site_id, :sid), " +
				"current_release_id = :rid, " +
				"suspended = :z, " +
				"created_at = if_not_exists(created_at, :t0), " +
				"updated_at = :t",
		),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":sid": avS(id.String()),
			":rid": avS(releaseID.String()),
			":z":   avN(int64(0)),
			":t0":  avN(now),
			":t":   avN(now),
		},
	})

	return err
}

func decodeSite(m map[string]types.AttributeValue) domain.Site {
	suspended := getN(m, "suspended") > 0
	return domain.Site{
		ID:             domain.SiteID(getS(m, "site_id")),
		Host:           getS(m, "host"),
		CurrentRelease: domain.ReleaseID(getS(m, "current_release_id")),
		Suspended:      suspended,
		CreatedAt:      time.Unix(getN(m, "created_at"), 0).UTC(),
		UpdatedAt:      time.Unix(getN(m, "updated_at"), 0).UTC(),
	}
}

var _ ports.SiteStore = (*SiteStore)(nil)
