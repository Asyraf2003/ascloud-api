package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func (s *SessionStore) isCurrent(ctx context.Context, userID, sid string) (bool, error) {
	out, err := s.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      &s.table,
		Key:            map[string]types.AttributeValue{"pk": avS(usrPK(userID))},
		ConsistentRead: awsBool(true),
	})
	if err != nil {
		return false, err
	}
	if out.Item == nil {
		return false, nil
	}
	return getS(out.Item, "current_sid") == sid, nil
}
