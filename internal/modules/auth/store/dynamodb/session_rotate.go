package dynamodb

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"

	"example.com/your-api/internal/modules/auth/domain"
)

func (s *SessionStore) RotateRefreshTokenHash(ctx context.Context, sessionID, oldHash, newHash string, newExpiresAt time.Time) error {
	sessOut, err := s.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      &s.table,
		Key:            map[string]types.AttributeValue{"pk": avS(sidPK(sessionID))},
		ConsistentRead: awsBool(true),
	})
	if err != nil {
		return err
	}
	if sessOut.Item == nil {
		return domain.ErrSessionNotFound
	}
	uid := getS(sessOut.Item, "user_id")
	if ok, err := s.isCurrent(ctx, uid, sessionID); err != nil {
		return err
	} else if !ok {
		return domain.ErrSessionRevoked
	}

	now := time.Now().UTC().Unix()
	exp := newExpiresAt.UTC().Unix()

	_, err = s.db.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{Update: &types.Update{
				TableName:           &s.table,
				Key:                 map[string]types.AttributeValue{"pk": avS(sidPK(sessionID))},
				ConditionExpression: awsStr("refresh_cur = :old AND attribute_not_exists(revoked_at) AND expires_at >= :now"),
				UpdateExpression:    awsStr("SET refresh_prev = refresh_cur, refresh_cur = :new, expires_at = :exp, ttl = :exp, updated_at = :now"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":old": avS(oldHash), ":new": avS(newHash), ":exp": avN(exp), ":now": avN(now),
				},
			}},
			{Update: &types.Update{
				TableName:           &s.table,
				Key:                 map[string]types.AttributeValue{"pk": avS(rhPK(oldHash))},
				ConditionExpression: awsStr("sid = :sid AND kind = :cur"),
				UpdateExpression:    awsStr("SET kind = :prev, updated_at = :now"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":sid": avS(sessionID), ":cur": avS("cur"), ":prev": avS("prev"), ":now": avN(now),
				},
			}},
			{Put: &types.Put{
				TableName: &s.table,
				Item: map[string]types.AttributeValue{
					"pk": avS(rhPK(newHash)), "sid": avS(sessionID), "user_id": avS(uid), "kind": avS("cur"), "created_at": avN(now),
				},
				ConditionExpression: awsStr("attribute_not_exists(pk)"),
			}},
		},
	})
	if err == nil {
		return nil
	}

	var tce *types.TransactionCanceledException
	if errors.As(err, &tce) {
		return domain.ErrRefreshTokenReused
	}
	var cfe *types.ConditionalCheckFailedException
	if errors.As(err, &cfe) {
		return domain.ErrRefreshTokenReused
	}
	return err
}
