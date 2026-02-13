package dynamodb

import (
	"context"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

type Config struct {
	Region   string
	Endpoint string // optional, e.g. http://localhost:8000 for DynamoDB Local
}

func ConfigFromEnv(defRegion string) Config {
	region := strings.TrimSpace(os.Getenv("AWS_REGION"))
	if region == "" {
		region = strings.TrimSpace(os.Getenv("AWS_DEFAULT_REGION"))
	}
	if region == "" {
		region = defRegion
	}
	return Config{Region: region, Endpoint: strings.TrimSpace(os.Getenv("DYNAMODB_ENDPOINT"))}
}

func New(ctx context.Context, cfg Config) (*dynamodb.Client, error) {
	opts := []func(*awscfg.LoadOptions) error{awscfg.WithRegion(cfg.Region)}
	if cfg.Endpoint != "" {
		creds := credentials.NewStaticCredentialsProvider("local", "local", "")
		opts = append(opts, awscfg.WithCredentialsProvider(creds))
		opts = append(opts, awscfg.WithEndpointResolverWithOptions(
			aws.EndpointResolverWithOptionsFunc(func(service, region string, _ ...any) (aws.Endpoint, error) {
				if service == dynamodb.ServiceID {
					return aws.Endpoint{URL: cfg.Endpoint, SigningRegion: cfg.Region, HostnameImmutable: true}, nil
				}
				return aws.Endpoint{}, &aws.EndpointNotFoundError{}
			}),
		))
	}
	awsCfg, err := awscfg.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(awsCfg), nil
}
