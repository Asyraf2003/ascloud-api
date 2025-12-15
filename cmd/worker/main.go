package main

import (
	"example.com/your-api/internal/config"
	"example.com/your-api/internal/shared/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(cfg.Env).With("service", "worker")
	log.Info("worker_stub", "status", "not_implemented_yet")
}
