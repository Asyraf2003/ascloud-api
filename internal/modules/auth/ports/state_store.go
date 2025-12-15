package ports

import (
	"context"
	"time"
)

type AuthState struct {
	Nonce        string
	CodeVerifier string
	Purpose      string // login | stepup
	CreatedAt    time.Time
}

type AuthStateStore interface {
	Put(ctx context.Context, state string, v AuthState, ttl time.Duration) error

	// GetDel: baca lalu hapus (anti-replay).
	GetDel(ctx context.Context, state string) (AuthState, error)
}
