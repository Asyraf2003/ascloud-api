package main

import (
	"context"

	hostingwire "example.com/your-api/internal/modules/hosting/wire"
	ddbplat "example.com/your-api/internal/platform/datastore/dynamodb"
	s3plat "example.com/your-api/internal/platform/objectstore/s3"
)

const defRegion = "ap-southeast-3"

func buildDeployer(ctx context.Context) (deployer, error) {
	ddb, err := ddbplat.New(ctx, ddbplat.ConfigFromEnv(defRegion))
	if err != nil {
		return nil, err
	}
	s3, err := s3plat.New(ctx, s3plat.ConfigFromEnv(defRegion))
	if err != nil {
		return nil, err
	}
	return hostingwire.WireHostingDeployer(ddb, s3)
}
