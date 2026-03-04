// TODO justify: >100 lines sementara; akan dipecah saat milestone 9.
package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambdacontext"

	"example.com/your-api/internal/modules/hosting/ports"
	"example.com/your-api/internal/shared/apperr"
	"example.com/your-api/internal/shared/obs"
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
	start := time.Now()

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

	// queue lag (ms) from enqueue time (if present)
	var queueLagMs float64
	if msg.QueuedAtUnix > 0 {
		queueLagMs = float64(time.Since(time.Unix(msg.QueuedAtUnix, 0)).Milliseconds())
	}

	log.Info("deploy_start")
	err := h.dep.Deploy(ctx, msg)

	durMs := float64(time.Since(start).Milliseconds())

	if err == nil {
		obs.EmitEMF(log, "your-api/hosting", map[string]string{
			"service": "worker",
			"op":      "deploy",
			"result":  "success",
			"code":    "ok",
		}, map[string]obs.MetricValue{
			"deploy_attempt_total": {Value: 1, Unit: "Count"},
			"deploy_success_total": {Value: 1, Unit: "Count"},
			"deploy_duration_ms":   {Value: durMs, Unit: "Milliseconds"},
			"queue_lag_ms":         {Value: queueLagMs, Unit: "Milliseconds"},
		})

		log.Info("deploy_ok")
		return false, ""
	}

	ae, ok := apperr.As(err)
	if !ok {
		code = "hosting.internal_error"

		obs.EmitEMF(log, "your-api/hosting", map[string]string{
			"service": "worker",
			"op":      "deploy",
			"result":  "fail",
			"code":    code,
		}, map[string]obs.MetricValue{
			"deploy_attempt_total": {Value: 1, Unit: "Count"},
			"deploy_fail_total":    {Value: 1, Unit: "Count"},
			"deploy_duration_ms":   {Value: durMs, Unit: "Milliseconds"},
			"queue_lag_ms":         {Value: queueLagMs, Unit: "Milliseconds"},
		})

		log.Error("deploy_failed", "code", code, "err", err, "permanent", false)
		return true, code
	}

	code = ae.Code
	permanent := isPermanent(ae.Code)

	obs.EmitEMF(log, "your-api/hosting", map[string]string{
		"service": "worker",
		"op":      "deploy",
		"result":  "fail",
		"code":    code,
	}, map[string]obs.MetricValue{
		"deploy_attempt_total": {Value: 1, Unit: "Count"},
		"deploy_fail_total":    {Value: 1, Unit: "Count"},
		"deploy_duration_ms":   {Value: durMs, Unit: "Milliseconds"},
		"queue_lag_ms":         {Value: queueLagMs, Unit: "Milliseconds"},
	})

	log.Error("deploy_failed", "code", code, "err", err, "permanent", permanent)

	if permanent {
		return false, code
	}
	return true, code
}

func lambdaRequestID(ctx context.Context) string {
	lc, ok := lambdacontext.FromContext(ctx)
	if !ok || lc == nil {
		return ""
	}
	return lc.AwsRequestID
}
