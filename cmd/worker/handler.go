package main

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"

	"example.com/your-api/internal/modules/hosting/ports"
	"example.com/your-api/internal/shared/apperr"
	"example.com/your-api/internal/shared/requestid"
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
	lambdaID := lambdaRequestID(ctx)
	h.log.Info("sqs_batch_start", "lambda_request_id", lambdaID, "records", len(evt.Records))

	var resp sqsBatchResponse
	for _, r := range evt.Records {
		retry, code := h.handleRecord(ctx, lambdaID, r)
		if !retry {
			continue
		}
		resp.BatchItemFailures = append(resp.BatchItemFailures, sqsBatchItemFailure{
			ItemIdentifier: r.MessageId,
		})
		h.log.Error("sqs_item_failed", "lambda_request_id", lambdaID, "msg_id", r.MessageId, "code", code)
	}

	h.log.Info("sqs_batch_done", "lambda_request_id", lambdaID, "failed", len(resp.BatchItemFailures))
	return resp, nil
}

func (h *sqsHandler) handleRecord(ctx context.Context, lambdaID string, r events.SQSMessage) (retry bool, code string) {
	var msg ports.DeployMessage
	if err := json.Unmarshal([]byte(r.Body), &msg); err != nil {
		h.log.Error("sqs_bad_message", "lambda_request_id", lambdaID, "msg_id", r.MessageId, "err", err)
		return false, "hosting.bad_message"
	}

	// propagate request_id into context for downstream visibility
	ctx = requestid.With(ctx, msg.RequestID)

	log := h.log.With(
		"lambda_request_id", lambdaID,
		"request_id", msg.RequestID,
		"site_id", msg.SiteID,
		"upload_id", msg.UploadID,
		"release_id", msg.ReleaseID,
		"msg_id", r.MessageId,
	)

	log.Info("deploy_start")
	err := h.dep.Deploy(ctx, msg)
	if err == nil {
		log.Info("deploy_ok")
		return false, ""
	}

	ae, ok := apperr.As(err)
	if !ok {
		log.Error("deploy_failed", "code", "hosting.internal_error", "err", err, "permanent", false)
		return true, "hosting.internal_error"
	}

	if isPermanent(ae.Code) {
		log.Error("deploy_failed", "code", ae.Code, "err", err, "permanent", true)
		return false, ae.Code
	}

	log.Error("deploy_failed", "code", ae.Code, "err", err, "permanent", false)
	return true, ae.Code
}

func lambdaRequestID(ctx context.Context) string {
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok || lc == nil {
		return ""
	}
	return lc.AwsRequestID
}
