package dynamodb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"example.com/your-api/internal/modules/auth/ports"
)

type AccountService struct {
	db    *dynamodb.Client
	table string
}

func NewAccountService(db *dynamodb.Client, table string) *AccountService {
	return &AccountService{db: db, table: table}
}

func (s *AccountService) Create(ctx context.Context, in ports.AccountInput) (uuid.UUID, error) {
	id := uuid.New()
	now := time.Now().UTC().Unix()
	meta, err := jsonOrEmpty(in.Meta)
	if err != nil {
		return uuid.Nil, err
	}

	item := map[string]types.AttributeValue{
		"pk":         avS(accPK(id.String())),
		"account_id": avS(id.String()),
		"email":      avS(in.Email),
		"created_at": avN(now),
		"updated_at": avN(now),
		"meta_json":  avS(meta),
	}
	_, err = s.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &s.table, Item: item,
		ConditionExpression: awsStr("attribute_not_exists(pk)"),
	})
	if err != nil {
		return uuid.Nil, err
	}
	return id, nil
}

var _ ports.AccountService = (*AccountService)(nil)
