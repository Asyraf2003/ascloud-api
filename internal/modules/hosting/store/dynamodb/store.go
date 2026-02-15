package dynamodb

import (
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"example.com/your-api/internal/modules/hosting/ports"
)

type UploadStore struct {
	db    *dynamodb.Client
	table string
}

func NewUploadStore(db *dynamodb.Client, table string) *UploadStore {
	return &UploadStore{db: db, table: table}
}

var _ ports.UploadStore = (*UploadStore)(nil)
