package aws

import (
	"context"
	"io"
	"strings"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	s3svc "github.com/aws/aws-sdk-go-v2/service/s3"
)

func (o *ObjectStore) Get(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	key := strings.TrimSpace(objectKey)
	out, err := o.s3.GetObject(ctx, &s3svc.GetObjectInput{
		Bucket: awsv2.String(o.bucket),
		Key:    awsv2.String(key),
	})
	if err != nil {
		return nil, err
	}
	return out.Body, nil
}

func (o *ObjectStore) Put(ctx context.Context, objectKey string, body io.Reader, contentType string, cacheControl string) error {
	key := strings.TrimSpace(objectKey)

	in := &s3svc.PutObjectInput{
		Bucket: awsv2.String(o.bucket),
		Key:    awsv2.String(key),
		Body:   body,
	}
	if strings.TrimSpace(contentType) != "" {
		in.ContentType = awsv2.String(strings.TrimSpace(contentType))
	}
	if strings.TrimSpace(cacheControl) != "" {
		in.CacheControl = awsv2.String(strings.TrimSpace(cacheControl))
	}

	_, err := o.s3.PutObject(ctx, in)
	return err
}

func (o *ObjectStore) Delete(ctx context.Context, objectKey string) error {
	key := strings.TrimSpace(objectKey)
	_, err := o.s3.DeleteObject(ctx, &s3svc.DeleteObjectInput{
		Bucket: awsv2.String(o.bucket),
		Key:    awsv2.String(key),
	})
	return err
}
