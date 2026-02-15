package aws

import (
	"context"
	"encoding/json"
	"strings"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	sqssvc "github.com/aws/aws-sdk-go-v2/service/sqs"

	"example.com/your-api/internal/modules/hosting/ports"
)

type DeployQueue struct {
	queueURL string
	sqs      *sqssvc.Client
}

func NewDeployQueue(sqs *sqssvc.Client, queueURL string) *DeployQueue {
	return &DeployQueue{sqs: sqs, queueURL: strings.TrimSpace(queueURL)}
}

func (q *DeployQueue) EnqueueDeploy(ctx context.Context, msg ports.DeployMessage) error {
	b, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	_, err = q.sqs.SendMessage(ctx, &sqssvc.SendMessageInput{
		QueueUrl:    awsv2.String(q.queueURL),
		MessageBody: awsv2.String(string(b)),
	})
	return err
}

var _ ports.DeployQueue = (*DeployQueue)(nil)
