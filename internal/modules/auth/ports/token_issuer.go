package ports

import (
	"context"
	"time"
)

type AccessTokenRequest struct {
	AccountID  string
	SessionID  string
	TrustLevel string // aal1|aal2
}

type TokenIssuer interface {
	IssueAccessToken(ctx context.Context, req AccessTokenRequest) (token string, exp time.Time, err error)
}
