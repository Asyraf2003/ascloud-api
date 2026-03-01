package main

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"

	"example.com/your-api/internal/modules/hosting/ports"
	"example.com/your-api/internal/shared/apperr"
)

type fakeDeployer struct {
	err error
}

func (f fakeDeployer) Deploy(context.Context, ports.DeployMessage) error { return f.err }

func TestHandler_PermanentDoesNotRetry(t *testing.T) {
	h := newSQSHandler(nil, fakeDeployer{err: apperr.New("hosting.zip_slip", 0, "")})
	resp, _ := h.Handle(context.Background(), events.SQSEvent{
		Records: []events.SQSMessage{{
			MessageId: "m1",
			Body:      `{"site_id":"s","upload_id":"u","release_id":"r","object_key":"k","size_bytes":1,"queued_at_unix":1}`,
		}},
	})
	if len(resp.BatchItemFailures) != 0 {
		t.Fatalf("want no failures, got %+v", resp.BatchItemFailures)
	}
}

func TestHandler_TransientRetries(t *testing.T) {
	h := newSQSHandler(nil, fakeDeployer{err: apperr.Wrap(errors.New("x"), "hosting.s3_get_failed", 0, "")})
	resp, _ := h.Handle(context.Background(), events.SQSEvent{
		Records: []events.SQSMessage{{
			MessageId: "m1",
			Body:      `{"site_id":"s","upload_id":"u","release_id":"r","object_key":"k","size_bytes":1,"queued_at_unix":1}`,
		}},
	})
	if len(resp.BatchItemFailures) != 1 || resp.BatchItemFailures[0].ItemIdentifier != "m1" {
		t.Fatalf("want retry m1, got %+v", resp.BatchItemFailures)
	}
}
