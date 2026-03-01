package usecase

import (
	"context"
	"time"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
)

func (d *Deployer) initRelease(ctx context.Context, msg ports.DeployMessage) {
	now := time.Now().UTC()
	_ = d.rel.Put(ctx, domain.Release{
		ID:        msg.ReleaseID,
		SiteID:    msg.SiteID,
		Status:    domain.ReleaseStatusPending,
		SizeBytes: 0,
		CreatedAt: now,
		UpdatedAt: now,
	})
}
