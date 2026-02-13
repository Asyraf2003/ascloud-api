package dynamodb

import "github.com/aws/aws-sdk-go-v2/aws"

func awsStr(s string) *string { return aws.String(s) }
