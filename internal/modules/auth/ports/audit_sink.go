package ports

import (
	"context"
	"time"
)

type AuditEvent struct {
	AccountID string // boleh kosong untuk attempt yang gagal sebelum resolve
	Event     string
	At        time.Time
	Meta      map[string]any // wajib allowlist/redact sebelum log
}

type AuditSink interface {
	Record(ctx context.Context, e AuditEvent) error
}
