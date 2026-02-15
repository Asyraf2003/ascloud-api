package usecase

import "example.com/your-api/internal/modules/hosting/ports"

type Service struct {
	cfg   Config
	up    ports.UploadStore
	obj   ports.ObjectStore
	queue ports.DeployQueue
}

func New(cfg Config, up ports.UploadStore, obj ports.ObjectStore, q ports.DeployQueue) *Service {
	return &Service{cfg: cfg, up: up, obj: obj, queue: q}
}
