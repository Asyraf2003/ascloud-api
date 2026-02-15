package wire

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	s3svc "github.com/aws/aws-sdk-go-v2/service/s3"
	sqssvc "github.com/aws/aws-sdk-go-v2/service/sqs"

	hostingaws "example.com/your-api/internal/modules/hosting/adapters/aws"
	hostingddb "example.com/your-api/internal/modules/hosting/store/dynamodb"
	hostinguc "example.com/your-api/internal/modules/hosting/usecase"
)

func envTrim(k, def string) string {
	v := strings.TrimSpace(os.Getenv(k))
	if v == "" {
		return def
	}
	return v
}

// WireHostingUploadPipeline wires Step 4 hosting upload pipeline dependencies:
// DynamoDB UploadStore + S3 ObjectStore + SQS DeployQueue.
func WireHostingUploadPipeline(ddb *dynamodb.Client, s3 *s3svc.Client, sqs *sqssvc.Client) (*hostinguc.Service, error) {
	if ddb == nil || s3 == nil || sqs == nil {
		return nil, fmt.Errorf("hosting wire: nil client (ddb=%v s3=%v sqs=%v)", ddb != nil, s3 != nil, sqs != nil)
	}

	table := envTrim("DDB_HOSTING_UPLOADS_TABLE", "hosting_uploads")
	bucket := envTrim("HOSTING_S3_BUCKET", "")
	queueURL := envTrim("HOSTING_SQS_QUEUE_URL", "")

	if bucket == "" {
		return nil, fmt.Errorf("hosting wire: missing HOSTING_S3_BUCKET")
	}
	if queueURL == "" {
		return nil, fmt.Errorf("hosting wire: missing HOSTING_SQS_QUEUE_URL")
	}

	up := hostingddb.NewUploadStore(ddb, table)
	obj := hostingaws.NewObjectStore(s3, bucket)
	q := hostingaws.NewDeployQueue(sqs, queueURL)

	svc := hostinguc.New(hostinguc.DefaultConfig(), up, obj, q)
	return svc, nil
}
