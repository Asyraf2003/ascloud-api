package usecase

import "example.com/your-api/internal/modules/hosting/ports"

type Deployer struct {
	cfg   DeployConfig
	up    ports.UploadStore
	sites ports.SiteStore
	rel   ports.ReleaseStore
	obj   ports.ObjectStore
	audit ports.AuditSink
}

func NewDeployer(cfg DeployConfig, up ports.UploadStore, sites ports.SiteStore, rel ports.ReleaseStore, obj ports.ObjectStore, audit ports.AuditSink) *Deployer {
	return &Deployer{cfg: cfg, up: up, sites: sites, rel: rel, obj: obj, audit: audit}
}
