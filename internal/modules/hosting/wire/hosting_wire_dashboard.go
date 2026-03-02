package wire

import (
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	hostingaws "example.com/your-api/internal/modules/hosting/adapters/aws"
	hostingddb "example.com/your-api/internal/modules/hosting/store/dynamodb"
	hostinguc "example.com/your-api/internal/modules/hosting/usecase"
)

func WireHostingDashboard(ddb *dynamodb.Client, kvs *cloudfrontkeyvaluestore.Client) (*hostinguc.Dashboard, error) {
	if ddb == nil || kvs == nil {
		return nil, fmt.Errorf("hosting wire dashboard: nil client (ddb=%v kvs=%v)", ddb != nil, kvs != nil)
	}

	sitesTable := envTrim("DDB_HOSTING_SITES_TABLE", "hosting_sites")
	relsTable := envTrim("DDB_HOSTING_RELEASES_TABLE", "hosting_releases")
	baseDomain := envTrim("HOSTING_BASE_DOMAIN", "asyrafcloud.my.id")
	kvsARN := strings.TrimSpace(os.Getenv("HOSTING_CLOUDFRONT_KVS_ARN"))
	if kvsARN == "" {
		return nil, fmt.Errorf("hosting wire dashboard: missing HOSTING_CLOUDFRONT_KVS_ARN")
	}

	sites := hostingddb.NewSiteStore(ddb, sitesTable)
	rels := hostingddb.NewReleaseStore(ddb, relsTable)

	edge, err := hostingaws.NewCloudFrontKVS(kvs, kvsARN)
	if err != nil {
		return nil, err
	}

	return hostinguc.NewDashboard(hostinguc.DashboardConfig{BaseDomain: baseDomain}, sites, rels, edge), nil
}
