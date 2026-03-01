package usecase

import (
	"time"

	"example.com/your-api/internal/modules/hosting/usecase/zipsec"
)

type DeployConfig struct {
	MaxZipBytes   int64
	Extract       zipsec.Options
	CacheControl  string
	TmpDir        string
	DeployTimeout time.Duration
}

func DefaultDeployConfig() DeployConfig {
	return DeployConfig{
		MaxZipBytes:   20 * 1024 * 1024,
		Extract:       zipsec.DefaultOptions(),
		CacheControl:  "public, max-age=31536000, immutable",
		TmpDir:        "/tmp",
		DeployTimeout: 2 * time.Minute,
	}
}
