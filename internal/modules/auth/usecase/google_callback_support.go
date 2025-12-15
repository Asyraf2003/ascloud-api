package usecase

import (
	"context"
	"time"

	"example.com/your-api/internal/modules/auth/ports"
)

func audit0(u *GoogleFlow, ctx context.Context, accountID, event string, meta map[string]any) {
	_ = u.audit.Record(ctx, ports.AuditEvent{AccountID: accountID, Event: event, At: time.Now(), Meta: meta})
}

func trustDecision(u *GoogleFlow, ctx context.Context, accountID, purpose string, c ClientInfo) (ports.TrustDecision, error) {
	return u.trust.Evaluate(ctx, ports.TrustSignals{
		AccountID: accountID,
		Purpose:   purpose,
		IP:        deref(c.IPPrefix),
		UserAgent: c.UserAgentHash,
	})
}
