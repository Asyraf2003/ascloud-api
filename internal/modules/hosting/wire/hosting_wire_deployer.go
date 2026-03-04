package wire

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	s3svc "github.com/aws/aws-sdk-go-v2/service/s3"

	hostingaws "example.com/your-api/internal/modules/hosting/adapters/aws"
	hostingddb "example.com/your-api/internal/modules/hosting/store/dynamodb"
	hostinguc "example.com/your-api/internal/modules/hosting/usecase"
)

func WireHostingDeployer(ddb *dynamodb.Client, s3 *s3svc.Client) (*hostinguc.Deployer, error) {
	if ddb == nil || s3 == nil {
		return nil, fmt.Errorf("hosting wire: nil client (ddb=%v s3=%v)", ddb != nil, s3 != nil)
	}

	uploadsTable := envTrim("DDB_HOSTING_UPLOADS_TABLE", "hosting_uploads")
	sitesTable := envTrim("DDB_HOSTING_SITES_TABLE", "hosting_sites")
	releasesTable := envTrim("DDB_HOSTING_RELEASES_TABLE", "hosting_releases")
	auditTable := envTrim("DDB_AUDIT_EVENTS_TABLE", "audit_events")

	bucket := envTrim("HOSTING_S3_BUCKET", "")
	if bucket == "" {
		return nil, fmt.Errorf("hosting wire: missing HOSTING_S3_BUCKET")
	}

	up := hostingddb.NewUploadStore(ddb, uploadsTable)
	sites := hostingddb.NewSiteStore(ddb, sitesTable)
	rel := hostingddb.NewReleaseStore(ddb, releasesTable)
	obj := hostingaws.NewObjectStore(s3, bucket)
	audit := hostingddb.NewAuditSink(ddb, auditTable)

	d := hostinguc.NewDeployer(hostinguc.DefaultDeployConfig(), up, sites, rel, obj, audit)
	return d, nil
}
