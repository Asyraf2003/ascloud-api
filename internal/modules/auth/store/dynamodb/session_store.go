package dynamodb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"example.com/your-api/internal/modules/auth/domain"
	"example.com/your-api/internal/modules/auth/ports"
)

type SessionStore struct {
	db    *dynamodb.Client
	table string
}

func NewSessionStore(db *dynamodb.Client, table string) *SessionStore {
	return &SessionStore{db: db, table: table}
}

var _ ports.SessionStore = (*SessionStore)(nil)

func (s *SessionStore) Create(ctx context.Context, sess domain.Session) (domain.Session, error) {
	now := time.Now().UTC()
	if sess.ID == "" {
		sess.ID = uuid.NewString()
	}
	if sess.CreatedAt.IsZero() {
		sess.CreatedAt = now
	}
	if sess.ExpiresAt.IsZero() {
		sess.ExpiresAt = now.Add(24 * time.Hour)
	}

	meta, err := jsonOrEmpty(sess.Meta)
	if err != nil {
		return domain.Session{}, err
	}

	sItem := map[string]types.AttributeValue{
		"pk":          avS(sidPK(sess.ID)),
		"session_id":  avS(sess.ID),
		"user_id":     avS(sess.UserID),
		"refresh_cur": avS(sess.RefreshTokenHash),
		"device_id":   avS(sess.DeviceID),
		"ua_hash":     avS(sess.UserAgentHash),
		"created_at":  avN(sess.CreatedAt.UTC().Unix()),
		"expires_at":  avN(sess.ExpiresAt.UTC().Unix()),
		"ttl":         avN(sess.ExpiresAt.UTC().Unix()),
		"meta_json":   avS(meta),
	}
	if sess.IPPrefix != nil {
		sItem["ip_prefix"] = avS(*sess.IPPrefix)
	}
	if sess.ProjectID != nil {
		sItem["project_id"] = avS(*sess.ProjectID)
	}
	if sess.RevokedAt != nil {
		sItem["revoked_at"] = avN(sess.RevokedAt.UTC().Unix())
	}

	rItem := map[string]types.AttributeValue{
		"pk":      avS(rhPK(sess.RefreshTokenHash)),
		"sid":     avS(sess.ID),
		"user_id": avS(sess.UserID),
		"kind":    avS("cur"),
	}

	pItem := map[string]types.AttributeValue{
		"pk":          avS(usrPK(sess.UserID)),
		"current_sid": avS(sess.ID),
		"updated_at":  avN(now.Unix()),
	}

	_, err = s.db.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{Put: &types.Put{TableName: &s.table, Item: sItem, ConditionExpression: awsStr("attribute_not_exists(pk)")}},
			{Put: &types.Put{TableName: &s.table, Item: rItem, ConditionExpression: awsStr("attribute_not_exists(pk)")}},
			{Put: &types.Put{TableName: &s.table, Item: pItem}},
		},
	})
	if err != nil {
		return domain.Session{}, err
	}
	return sess, nil
}
