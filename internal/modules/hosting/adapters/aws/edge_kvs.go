// TODO justify: >100 lines sementara; akan dipecah saat milestone 9.

package aws

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync/atomic"

	awsv2 "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfrontkeyvaluestore"

	"example.com/your-api/internal/modules/hosting/domain"
	"example.com/your-api/internal/modules/hosting/ports"
)

type CloudFrontKVS struct {
	client *cloudfrontkeyvaluestore.Client
	kvsARN string
	etag   atomic.Value // string
}

func NewCloudFrontKVS(client *cloudfrontkeyvaluestore.Client, kvsARN string) (*CloudFrontKVS, error) {
	if client == nil {
		return nil, fmt.Errorf("cloudfront kvs: nil client")
	}
	kvsARN = strings.TrimSpace(kvsARN)
	if kvsARN == "" {
		return nil, fmt.Errorf("cloudfront kvs: empty kvs arn")
	}
	return &CloudFrontKVS{client: client, kvsARN: kvsARN}, nil
}

type hostMapping struct {
	SiteID         string `json:"site_id"`
	CurrentRelease string `json:"current_release_id"`
	Suspended      bool   `json:"suspended"`
}

func (k *CloudFrontKVS) PutHostMapping(ctx context.Context, host string, siteID domain.SiteID, currentReleaseID domain.ReleaseID, suspended bool) error {
	key := normalizeHost(host)
	if key == "" {
		return fmt.Errorf("cloudfront kvs: empty host key")
	}

	val := hostMapping{
		SiteID:         strings.TrimSpace(siteID.String()),
		CurrentRelease: strings.TrimSpace(currentReleaseID.String()),
		Suspended:      suspended,
	}
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}

	etag, err := k.getETag(ctx)
	if err != nil {
		return err
	}

	// Try once with cached ETag, if fails refresh ETag and retry once.
	if err := k.putKey(ctx, etag, key, string(b)); err == nil {
		return nil
	}

	etag2, err := k.refreshETag(ctx)
	if err != nil {
		return err
	}
	return k.putKey(ctx, etag2, key, string(b))
}

func (k *CloudFrontKVS) putKey(ctx context.Context, etag string, key string, value string) error {
	out, err := k.client.PutKey(ctx, &cloudfrontkeyvaluestore.PutKeyInput{
		IfMatch: awsv2.String(etag),
		Key:     awsv2.String(key),
		KvsARN:  awsv2.String(k.kvsARN),
		Value:   awsv2.String(value),
	})
	if err != nil {
		return err
	}
	if out != nil && out.ETag != nil && *out.ETag != "" {
		k.etag.Store(*out.ETag)
	}
	return nil
}

func (k *CloudFrontKVS) getETag(ctx context.Context) (string, error) {
	if v, ok := k.etag.Load().(string); ok && v != "" {
		return v, nil
	}
	return k.refreshETag(ctx)
}

func (k *CloudFrontKVS) refreshETag(ctx context.Context) (string, error) {
	out, err := k.client.DescribeKeyValueStore(ctx, &cloudfrontkeyvaluestore.DescribeKeyValueStoreInput{
		KvsARN: awsv2.String(k.kvsARN),
	})
	if err != nil {
		return "", err
	}
	if out == nil || out.ETag == nil || *out.ETag == "" {
		return "", fmt.Errorf("cloudfront kvs: missing etag from DescribeKeyValueStore")
	}
	k.etag.Store(*out.ETag)
	return *out.ETag, nil
}

func normalizeHost(host string) string {
	h := strings.TrimSpace(strings.ToLower(host))
	if h == "" {
		return ""
	}
	// strip :port
	if i := strings.IndexByte(h, ':'); i >= 0 {
		h = h[:i]
	}
	// strip trailing dot
	h = strings.TrimSuffix(h, ".")
	return strings.TrimSpace(h)
}

var _ ports.EdgeStore = (*CloudFrontKVS)(nil)
