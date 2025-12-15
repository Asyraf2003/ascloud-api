package ports

import (
	"context"
	"time"
)

type OIDCAuthURLParams struct {
	State         string
	Nonce         string
	CodeChallenge string
	RedirectURL   string
	Purpose       string // login | stepup
}

type OIDCClaims struct {
	Provider      string
	Subject       string
	Email         string
	EmailVerified bool
	AuthTime      time.Time
}

type OIDCProvider interface {
	AuthCodeURL(p OIDCAuthURLParams) string

	// ExchangeAndVerify wajib melakukan verifikasi token (iss/aud/exp/nonce).
	ExchangeAndVerify(ctx context.Context, code, codeVerifier, redirectURL, nonce string) (OIDCClaims, error)
}
