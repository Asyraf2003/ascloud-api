package usecase

import "example.com/your-api/internal/modules/hosting/ports"

type Deployer struct {
	cfg   DeployConfig
	up    ports.UploadStore
	sites ports.SiteStore
	rel   ports.ReleaseStore
	obj   ports.ObjectStore
}

func NewDeployer(cfg DeployConfig, up ports.UploadStore, sites ports.SiteStore, rel ports.ReleaseStore, obj ports.ObjectStore) *Deployer {
	return &Deployer{cfg: cfg, up: up, sites: sites, rel: rel, obj: obj}
}
