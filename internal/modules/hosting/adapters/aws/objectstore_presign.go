package aws

import (
	"context"
	"strings"
	"time"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	s3svc "github.com/aws/aws-sdk-go-v2/service/s3"
)

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
