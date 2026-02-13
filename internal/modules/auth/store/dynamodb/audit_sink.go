package dynamodb

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"example.com/your-api/internal/modules/auth/ports"
)

type AuditSink struct {
	db    *dynamodb.Client
	table string
}

func NewAuditSink(db *dynamodb.Client, table string) *AuditSink {
	return &AuditSink{db: db, table: table}
}

func (s *AuditSink) Record(ctx context.Context, e ports.AuditEvent) error {
	meta, err := jsonOrEmpty(e.Meta)
	if err != nil {
		return err
	}
	acc := e.AccountID
	if acc == "" {
		acc = "anon"
	}
	sk := fmt.Sprintf("ts#%d#%s", e.At.UTC().Unix(), uuid.NewString())
	item := map[string]types.AttributeValue{
		"pk":        avS(accPK(acc)),
		"sk":        avS(sk),
		"event":     avS(e.Event),
		"at":        avN(e.At.UTC().Unix()),
		"meta_json": avS(meta),
	}
	_, err = s.db.PutItem(ctx, &dynamodb.PutItemInput{TableName: &s.table, Item: item})
	return err
}

var _ ports.AuditSink = (*AuditSink)(nil)
