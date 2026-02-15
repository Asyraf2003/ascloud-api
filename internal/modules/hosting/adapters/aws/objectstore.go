package aws

import (
	"context"
	"strings"
	"time"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
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

func (o *ObjectStore) PresignPutZip(ctx context.Context, objectKey string, _ int64, expires time.Duration) (string, error) {
	key := strings.TrimSpace(objectKey)
	in := &s3svc.PutObjectInput{
		Bucket:      awsv2.String(o.bucket),
		Key:         awsv2.String(key),
		ContentType: awsv2.String("application/zip"),
	}
	out, err := o.presign.PresignPutObject(ctx, in, func(po *s3svc.PresignOptions) {
		if expires > 0 {
			po.Expires = expires
		}
	})
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

func (o *ObjectStore) Head(ctx context.Context, objectKey string) (int64, error) {
	key := strings.TrimSpace(objectKey)
	out, err := o.s3.HeadObject(ctx, &s3svc.HeadObjectInput{
		Bucket: awsv2.String(o.bucket),
		Key:    awsv2.String(key),
	})
	if err != nil {
		return 0, err
	}
	if out.ContentLength == nil {
		return 0, nil
	}
	return *out.ContentLength, nil
}

var _ ports.ObjectStore = (*ObjectStore)(nil)
