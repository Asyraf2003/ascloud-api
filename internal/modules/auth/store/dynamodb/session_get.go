package dynamodb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"example.com/your-api/internal/modules/auth/domain"
)

func (s *SessionStore) GetByID(ctx context.Context, id string) (domain.Session, error) {
	out, err := s.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      &s.table,
		Key:            map[string]types.AttributeValue{"pk": avS(sidPK(id))},
		ConsistentRead: awsBool(true),
	})
	if err != nil {
		return domain.Session{}, err
	}
	if out.Item == nil {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	sess := decodeSession(out.Item)
	if sess.RevokedAt != nil {
		return domain.Session{}, domain.ErrSessionRevoked
	}
	if sess.ExpiresAt.Before(time.Now().UTC()) {
		return domain.Session{}, domain.ErrSessionExpired
	}
	if ok, err := s.isCurrent(ctx, sess.UserID, sess.ID); err != nil {
		return domain.Session{}, err
	} else if !ok {
		return domain.Session{}, domain.ErrSessionRevoked
	}
	return sess, nil
}

func decodeSession(m map[string]types.AttributeValue) domain.Session {
	sid := getS(m, "session_id")
	uid := getS(m, "user_id")
	proj := getS(m, "project_id")
	ip := getS(m, "ip_prefix")
	ex := time.Unix(getN(m, "expires_at"), 0).UTC()
	cr := time.Unix(getN(m, "created_at"), 0).UTC()
	var revoked *time.Time
	if has(m, "revoked_at") {
		t := time.Unix(getN(m, "revoked_at"), 0).UTC()
		revoked = &t
	}
	var ipPtr *string
	if ip != "" {
		ipPtr = &ip
	}
	var projPtr *string
	if proj != "" {
		projPtr = &proj
	}
	return domain.Session{
		ID: sid, UserID: uid, ProjectID: projPtr,
		RefreshTokenHash: getS(m, "refresh_cur"),
		DeviceID:         getS(m, "device_id"), UserAgentHash: getS(m, "ua_hash"), IPPrefix: ipPtr,
		CreatedAt: cr, ExpiresAt: ex, RevokedAt: revoked,
		Meta: map[string]any{},
	}
}
