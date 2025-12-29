package usecase

import (
	"errors"
	"time"

	"example.com/your-api/internal/modules/auth/ports"
)

type GoogleFlow struct {
	oidc     ports.OIDCProvider
	states   ports.AuthStateStore
	sessions ports.SessionStore
	tx       ports.Transactor

	ids      ports.IdentityRepository
	accounts ports.AccountService
	tokens   ports.TokenIssuer
	trust    ports.TrustEvaluator
	audit    ports.AuditSink

	stateTTL   time.Duration
	refreshTTL time.Duration
	hashSecret string
}

// Backward-compatible ctor (tests masih pakai ini).
func NewGoogleFlow(
	oidc ports.OIDCProvider,
	states ports.AuthStateStore,
	sessions ports.SessionStore,
	ids ports.IdentityRepository,
	accounts ports.AccountService,
	tokens ports.TokenIssuer,
	trust ports.TrustEvaluator,
	audit ports.AuditSink,
	stateTTL, refreshTTL time.Duration,
	hashSecret string,
) (*GoogleFlow, error) {
	return NewGoogleFlowWithTx(
		oidc, states, sessions, ports.NoopTransactor{},
		ids, accounts, tokens, trust, audit,
		stateTTL, refreshTTL, hashSecret,
	)
}

// Ctor untuk produksi: transactor di-inject dari wire (postgres).
func NewGoogleFlowWithTx(
	oidc ports.OIDCProvider,
	states ports.AuthStateStore,
	sessions ports.SessionStore,
	tx ports.Transactor,
	ids ports.IdentityRepository,
	accounts ports.AccountService,
	tokens ports.TokenIssuer,
	trust ports.TrustEvaluator,
	audit ports.AuditSink,
	stateTTL, refreshTTL time.Duration,
	hashSecret string,
) (*GoogleFlow, error) {
	if oidc == nil || states == nil || sessions == nil || tx == nil || ids == nil || accounts == nil || tokens == nil || trust == nil || audit == nil {
		return nil, errors.New("missing deps")
	}
	if stateTTL <= 0 || refreshTTL <= 0 {
		return nil, errors.New("ttl invalid")
	}
	if hashSecret == "" {
		return nil, errors.New("hash secret empty")
	}

	return &GoogleFlow{
		oidc: oidc, states: states, sessions: sessions, tx: tx,
		ids: ids, accounts: accounts,
		tokens: tokens, trust: trust, audit: audit,
		stateTTL: stateTTL, refreshTTL: refreshTTL, hashSecret: hashSecret,
	}, nil
}
