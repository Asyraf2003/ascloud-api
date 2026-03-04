package ports

import (
	"context"
	"time"

	"example.com/your-api/internal/modules/hosting/domain"
)

type AuditEvent struct {
	SiteID  domain.SiteID // wajib untuk hosting (pk = site#<id>)
	ActorID string        // optional (kalau nanti dari JWT)
	Event   string
	At      time.Time
	Meta    map[string]any // wajib allowlist/redact sebelum log/serialize
}

type AuditSink interface {
	Record(ctx context.Context, e AuditEvent) error
}
