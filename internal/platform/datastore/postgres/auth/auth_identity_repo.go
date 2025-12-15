package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"example.com/your-api/internal/modules/auth/ports"
)

type AuthIdentityRepo struct{ db *sql.DB }

func NewAuthIdentityRepo(db *sql.DB) *AuthIdentityRepo { return &AuthIdentityRepo{db: db} }

func (r *AuthIdentityRepo) FindAccountIDByIdentity(ctx context.Context, provider, subject string) (uuid.UUID, bool, error) {
	var id uuid.UUID
	err := r.db.QueryRowContext(ctx,
		`SELECT account_id FROM auth_identities WHERE provider=$1 AND subject=$2`, provider, subject,
	).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, false, nil
		}
		return uuid.Nil, false, err
	}
	return id, true, nil
}

func (r *AuthIdentityRepo) UpsertIdentity(ctx context.Context, accountID uuid.UUID, provider, subject, email string, emailVerified bool, meta map[string]any) error {
	if meta == nil {
		meta = map[string]any{}
	}
	b, err := json.Marshal(meta)
	if err != nil {
		return err
	}

	res, err := r.db.ExecContext(ctx, `
INSERT INTO auth_identities(provider, subject, account_id, email, email_verified, meta)
VALUES($1,$2,$3,$4,$5,$6)
ON CONFLICT (provider, subject)
DO UPDATE SET email=EXCLUDED.email, email_verified=EXCLUDED.email_verified, meta=EXCLUDED.meta, updated_at=now()
WHERE auth_identities.account_id = EXCLUDED.account_id
`, provider, subject, accountID, email, emailVerified, b)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return errors.New("identity already linked to different account")
	}
	return nil
}

var _ ports.IdentityRepository = (*AuthIdentityRepo)(nil)
