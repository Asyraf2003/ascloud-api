package dynamodb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func (s *SessionStore) Revoke(ctx context.Context, sessionID string, revokedAt time.Time) error {
	out, err := s.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      &s.table,
		Key:            map[string]types.AttributeValue{"pk": avS(sidPK(sessionID))},
		ConsistentRead: awsBool(true),
	})
	if err != nil {
		return err
	}
	if out.Item == nil {
		return nil
	}
	uid := getS(out.Item, "user_id")

	_, _ = s.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &s.table,
		Key:                       map[string]types.AttributeValue{"pk": avS(sidPK(sessionID))},
		UpdateExpression:          awsStr("SET revoked_at = :t"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":t": avN(revokedAt.UTC().Unix())},
	})

	_, _ = s.db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName:                 &s.table,
		Key:                       map[string]types.AttributeValue{"pk": avS(usrPK(uid))},
		ConditionExpression:       awsStr("current_sid = :sid"),
		UpdateExpression:          awsStr("REMOVE current_sid SET updated_at = :t"),
		ExpressionAttributeValues: map[string]types.AttributeValue{":sid": avS(sessionID), ":t": avN(time.Now().UTC().Unix())},
	})

	return nil
}
