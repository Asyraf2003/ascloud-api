package aws

import (
	"context"
	"strings"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	s3svc "github.com/aws/aws-sdk-go-v2/service/s3"
)

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
