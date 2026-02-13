package dynamodb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"example.com/your-api/internal/modules/auth/ports"
)

type IdentityRepo struct {
	db    *dynamodb.Client
	table string
}

func NewIdentityRepo(db *dynamodb.Client, table string) *IdentityRepo {
	return &IdentityRepo{db: db, table: table}
}

func (r *IdentityRepo) FindAccountIDByIdentity(ctx context.Context, provider, subject string) (uuid.UUID, bool, error) {
	out, err := r.db.GetItem(ctx, &dynamodb.GetItemInput{
		TableName:      &r.table,
		Key:            map[string]types.AttributeValue{"pk": avS(idpPK(provider, subject))},
		ConsistentRead: awsBool(true),
	})
	if err != nil {
		return uuid.Nil, false, err
	}
	if out.Item == nil {
		return uuid.Nil, false, nil
	}
	id, err := uuid.Parse(getS(out.Item, "account_id"))
	if err != nil {
		return uuid.Nil, false, err
	}
	return id, true, nil
}

func (r *IdentityRepo) UpsertIdentity(ctx context.Context, accountID uuid.UUID, provider, subject, email string, emailVerified bool, meta map[string]any) error {
	now := time.Now().UTC().Unix()
	mj, err := jsonOrEmpty(meta)
	if err != nil {
		return err
	}
	item := map[string]types.AttributeValue{
		"pk":             avS(idpPK(provider, subject)),
		"provider":       avS(provider),
		"subject":        avS(subject),
		"account_id":     avS(accountID.String()),
		"email":          avS(email),
		"email_verified": avS(boolStr(emailVerified)),
		"updated_at":     avN(now),
		"meta_json":      avS(mj),
	}
	_, err = r.db.PutItem(ctx, &dynamodb.PutItemInput{TableName: &r.table, Item: item})
	return err
}

var _ ports.IdentityRepository = (*IdentityRepo)(nil)
