package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"

	"example.com/your-api/internal/config"
	"example.com/your-api/internal/shared/logger"
)

func run() {
	cfg := config.Load()
	log := logger.New(cfg.Env).With("service", "worker")

	ctx := context.Background()
	dep, err := buildDeployer(ctx)
	if err != nil {
		log.Error("worker_boot_failed", "err", err)
		return
	}

	h := newSQSHandler(log, dep)
	lambda.Start(h.Handle)
}
