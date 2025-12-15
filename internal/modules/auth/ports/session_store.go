package ports

import (
	"context"
	"time"

	"example.com/your-api/internal/modules/auth/domain"
)

// SessionStore = kontrak penyimpanan session.
// Implementasi awal: Postgres.
// Implementasi nanti: Dynamo (tanpa ubah usecase).
type SessionStore interface {
	// Create menyimpan session baru dan mengembalikan session yang sudah punya ID+CreatedAt.
	Create(ctx context.Context, s domain.Session) (domain.Session, error)

	GetByID(ctx context.Context, id string) (domain.Session, error)
	GetByRefreshTokenHash(ctx context.Context, hash string) (domain.Session, error)

	// Rotate harus atomic dan menolak reuse.
	RotateRefreshTokenHash(ctx context.Context, sessionID string, oldHash string, newHash string, newExpiresAt time.Time) error

	// Revoke idempotent: kalau sudah revoked, tetap sukses.
	Revoke(ctx context.Context, sessionID string, revokedAt time.Time) error
}
