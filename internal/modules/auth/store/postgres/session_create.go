package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"example.com/your-api/internal/platform/datastore/postgres"

	"example.com/your-api/internal/modules/auth/domain"
)

func (s *SessionStorePostgres) Create(ctx context.Context, sess domain.Session) (domain.Session, error) {
	meta := sess.Meta
	if meta == nil {
		meta = map[string]any{}
	}
	metaJSON, err := json.Marshal(meta)
	if err != nil {
		return domain.Session{}, err
	}

	uid, err := postgres.ParseUUID(sess.UserID)
	if err != nil {
		return domain.Session{}, err
	}
	var pid any = nil
	if sess.ProjectID != nil && *sess.ProjectID != "" {
		p, err := postgres.ParseUUID(*sess.ProjectID)
		if err != nil {
			return domain.Session{}, err
		}
		pid = p
	}

	tx, err := s.db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return domain.Session{}, err
	}
	defer tx.Rollback()

	now := time.Now()
	if _, err := tx.ExecContext(ctx, `UPDATE auth_sessions SET revoked_at=$2 WHERE user_id=$1 AND revoked_at IS NULL`, uid, now); err != nil {
		return domain.Session{}, err
	}

	var id string
	var createdAt time.Time
	err = tx.QueryRowContext(ctx, qInsertSession,
		uid, pid,
		sess.RefreshTokenHash,
		sess.DeviceID, sess.UserAgentHash, sess.IPPrefix,
		sess.ExpiresAt, metaJSON,
	).Scan(&id, &createdAt)
	if err != nil {
		return domain.Session{}, err
	}
	if err := tx.Commit(); err != nil {
		return domain.Session{}, err
	}

	sess.ID = id
	sess.CreatedAt = createdAt
	sess.Meta = meta
	return sess, nil
}
