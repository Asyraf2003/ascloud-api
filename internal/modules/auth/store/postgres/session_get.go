package auth

import (
	"context"

	"example.com/your-api/internal/platform/datastore/postgres"

	"example.com/your-api/internal/modules/auth/domain"
)

func (s *SessionStorePostgres) GetByID(ctx context.Context, id string) (domain.Session, error) {
	uid, err := postgres.ParseUUID(id)
	if err != nil {
		return domain.Session{}, err
	}
	return scanSession(ctx, s.db.QueryRowContext(ctx, qSelectSessionByID, uid))
}

func (s *SessionStorePostgres) GetByRefreshTokenHash(ctx context.Context, hash string) (domain.Session, error) {
	return scanSession(ctx, s.db.QueryRowContext(ctx, qSelectSessionByRefreshHash, hash))
}
