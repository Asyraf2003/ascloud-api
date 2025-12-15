package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/google/uuid"

	"example.com/your-api/internal/modules/auth/ports"
)

type AuthAccountService struct {
	db *sql.DB
}

func NewAuthAccountService(db *sql.DB) *AuthAccountService { return &AuthAccountService{db: db} }

func (s *AuthAccountService) Create(ctx context.Context, in ports.AccountInput) (uuid.UUID, error) {
	if in.Email == "" {
		return uuid.Nil, errors.New("email empty")
	}
	meta := in.Meta
	if meta == nil {
		meta = map[string]any{}
	}
	b, err := json.Marshal(meta)
	if err != nil {
		return uuid.Nil, err
	}

	var id uuid.UUID
	err = s.db.QueryRowContext(ctx,
		`INSERT INTO accounts(email, meta) VALUES($1,$2) ON CONFLICT(email) DO UPDATE SET meta=EXCLUDED.meta RETURNING id`,
		in.Email, b,
	).Scan(&id)
	return id, err
}

var _ ports.AccountService = (*AuthAccountService)(nil)
