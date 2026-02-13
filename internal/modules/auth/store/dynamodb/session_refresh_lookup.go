package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"example.com/your-api/internal/modules/auth/domain"
)

func (s *SessionStore) GetByRefreshTokenHash(ctx context.Context, hash string) (domain.Session, error) {
	out, err := s.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      &s.table,
		Key:            map[string]types.AttributeValue{"pk": avS(rhPK(hash))},
		ConsistentRead: awsBool(true),
	})
	if err != nil {
		return domain.Session{}, err
	}
	if out.Item == nil {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	sid := getS(out.Item, "sid")
	if sid == "" {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	return s.GetByID(ctx, sid)
}
