package aws

import (
	"strings"

	s3svc "github.com/aws/aws-sdk-go-v2/service/s3"

	"example.com/your-api/internal/modules/hosting/ports"
)

type ObjectStore struct {
	bucket  string
	presign *s3svc.PresignClient
	s3      *s3svc.Client
}

func NewObjectStore(s3 *s3svc.Client, bucket string) *ObjectStore {
	b := strings.TrimSpace(bucket)
	return &ObjectStore{
		bucket:  b,
		s3:      s3,
		presign: s3svc.NewPresignClient(s3),
	}
}

var _ ports.ObjectStore = (*ObjectStore)(nil)
