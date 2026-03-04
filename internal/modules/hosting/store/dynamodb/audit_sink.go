package dynamodb

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"

	"example.com/your-api/internal/modules/hosting/ports"
	"example.com/your-api/internal/shared/redact"
)

type AuditSink struct {
	db    *dynamodb.Client
	table string
}

func NewAuditSink(db *dynamodb.Client, table string) *AuditSink {
	return &AuditSink{db: db, table: table}
}

func (s *AuditSink) Record(ctx context.Context, e ports.AuditEvent) error {
	pk := "site#unknown"
	if e.SiteID != "" {
		pk = "site#" + string(e.SiteID)
	}

	metaJSON, err := jsonOrEmpty(redact.RedactMap(e.Meta))
	if err != nil {
		return err
	}

	at := e.At.UTC()
	sk := fmt.Sprintf("ts#%d#%s", at.Unix(), uuid.NewString())

	item := map[string]types.AttributeValue{
		"pk":        &types.AttributeValueMemberS{Value: pk},
		"sk":        &types.AttributeValueMemberS{Value: sk},
		"event":     &types.AttributeValueMemberS{Value: e.Event},
		"at":        &types.AttributeValueMemberN{Value: strconv.FormatInt(at.Unix(), 10)},
		"meta_json": &types.AttributeValueMemberS{Value: metaJSON},
	}

	_, err = s.db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: &s.table,
		Item:      item,
	})
	return err
}

func jsonOrEmpty(m map[string]any) (string, error) {
	if len(m) == 0 {
		return "{}", nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

var _ ports.AuditSink = (*AuditSink)(nil)
