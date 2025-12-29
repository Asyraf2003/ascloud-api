package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"

	"example.com/your-api/internal/modules/auth/ports"
	"example.com/your-api/internal/platform/datastore/postgres" // Import platform
	"github.com/google/uuid"
)

type AuthAccountService struct {
	db *sql.DB // Simpan *sql.DB original
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

	// AMBIL EXECUTOR (Bisa TX atau DB biasa)
	executor := postgres.GetExecutor(ctx, s.db)

	var id uuid.UUID
	// Point penting: executor sekarang berisi *sql.Tx (jika dalam transaksi)
	// atau *sql.DB (jika transaksi tidak ada). Ini sangat AMAN.
	err = executor.QueryRowContext(ctx,
		`INSERT INTO accounts(email, meta) VALUES($1,$2) RETURNING id`,
		in.Email, b,
	).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return uuid.Nil, ports.ErrAccountEmailTaken
		}
		return uuid.Nil, err
	}

	return id, nil
}
