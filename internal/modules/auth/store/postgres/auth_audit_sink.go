package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"example.com/your-api/internal/modules/auth/ports"
	"example.com/your-api/internal/shared/redact"
)

type AuthAuditSink struct{ db *sql.DB }

func NewAuthAuditSink(db *sql.DB) *AuthAuditSink { return &AuthAuditSink{db: db} }

func (s *AuthAuditSink) Record(ctx context.Context, e ports.AuditEvent) error {
	at := e.At
	if at.IsZero() {
		at = time.Now()
	}
	meta := redact.RedactMap(e.Meta)
	b, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	var acc any = nil
	if e.AccountID != "" {
		if id, err := uuid.Parse(e.AccountID); err == nil {
			acc = id
		}
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT INTO auth_audit_events(id, account_id, event, at, meta) VALUES($1,$2,$3,$4,$5)`,
		uuid.New(), acc, e.Event, at, b,
	)
	return err
}

var _ ports.AuditSink = (*AuthAuditSink)(nil)
