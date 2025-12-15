package auth

import (
	"context"
	"database/sql"
	"time"

	"example.com/your-api/internal/modules/auth/domain"
)

func (s *SessionStorePostgres) RotateRefreshTokenHash(ctx context.Context, sessionID string, oldHash string, newHash string, newExpiresAt time.Time) error {
	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, `INSERT INTO auth_refresh_used(hash, session_id) VALUES($1,$2) ON CONFLICT DO NOTHING`,
		oldHash, sessionID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrRefreshTokenReused
	}

	res, err = tx.ExecContext(ctx, qRotateRefreshHash, newHash, newExpiresAt, sessionID, oldHash)
	if err != nil {
		return err
	}
	n, _ = res.RowsAffected()
	if n == 0 {
		return domain.ErrRefreshTokenReused
	}
	return tx.Commit()
}

func (s *SessionStorePostgres) Revoke(ctx context.Context, sessionID string, revokedAt time.Time) error {
	_, err := s.db.ExecContext(ctx, qRevokeSession, sessionID, revokedAt)
	return err
}
