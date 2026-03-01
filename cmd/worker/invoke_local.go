//go:build localinvoke

package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"

	"example.com/your-api/internal/config"
	"example.com/your-api/internal/shared/logger"
)

func main() {
	eventPath := flag.String("event", "", "path to SQSEvent json")
	flag.Parse()

	if *eventPath == "" {
		fmt.Fprintln(os.Stderr, "missing -event <path>")
		os.Exit(2)
	}

	b, err := os.ReadFile(*eventPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "read event:", err)
		os.Exit(1)
	}

	var evt events.SQSEvent
	if err := json.Unmarshal(b, &evt); err != nil {
		fmt.Fprintln(os.Stderr, "parse event:", err)
		os.Exit(1)
	}

	cfg := config.Load()
	log := logger.New(cfg.Env).With("service", "worker_localinvoke")

	dep, err := buildDeployer(context.Background())
	if err != nil {
		log.Error("boot_failed", "err", err)
		os.Exit(1)
	}

	h := newSQSHandler(log, dep)
	resp, err := h.Handle(context.Background(), evt)
	if err != nil {
		log.Error("handle_failed", "err", err)
		os.Exit(1)
	}

	out, _ := json.MarshalIndent(resp, "", "  ")
	fmt.Println(string(out))
}
