package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"

	"example.com/your-api/internal/modules/hosting/ports"
	"example.com/your-api/internal/shared/apperr"
)

type deployer interface {
	Deploy(ctx context.Context, msg ports.DeployMessage) error
}

type sqsHandler struct {
	log *slog.Logger
	dep deployer
}

func newSQSHandler(log *slog.Logger, dep deployer) *sqsHandler {
	if log == nil {
		log = slog.Default()
	}
	return &sqsHandler{log: log, dep: dep}
}

func (h *sqsHandler) Handle(ctx context.Context, evt events.SQSEvent) (sqsBatchResponse, error) {
	reqID := lambdaRequestID(ctx)
	h.log.Info("sqs_batch_start", "request_id", reqID, "records", len(evt.Records))

	var resp sqsBatchResponse
	for _, r := range evt.Records {
		retry, code := h.handleRecord(ctx, r)
		if !retry {
			continue
		}
		resp.BatchItemFailures = append(resp.BatchItemFailures, sqsBatchItemFailure{
			ItemIdentifier: r.MessageId,
		})
	}h.log.Info("sqs_batch_start", "lambda_request_id", lambdaID, "records", len(evt.Records))
	return resp, nil
}

func (h *sqsHandler) handleRecord(ctx context.Context, r events.SQSMessage) (retry bool, code string) {
	var msg ports.DeployMessage
	if err := json.Unmarshal([]byte(r.Body), &msg); err != nil {
		return false, "hosting.bad_message"
	}

	err := h.dep.Deploy(ctx, msg)
	if err == nil {
		return false, ""
	}

	ae, ok := apperr.As(err)
	if !ok {
		return true, "hosting.internal_error"
	}
	if isPermanent(ae.Code) {
		return false, ae.Code
	}
	return true, ae.Code
}

func lambdaRequestID(ctx context.Context) string {
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok || lc == nil {
		return ""
	}
	return lc.AwsRequestID
}
