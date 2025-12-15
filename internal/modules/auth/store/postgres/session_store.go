package auth

import (
	"database/sql"

	"example.com/your-api/internal/modules/auth/ports"
)

// SessionStorePostgres = implementasi SessionStore berbasis Postgres.
// Ini BLOK yang boleh diganti nanti (misal Dynamo) tanpa nyentuh usecase.
type SessionStorePostgres struct {
	db *sql.DB
}

var _ ports.SessionStore = (*SessionStorePostgres)(nil)

func NewSessionStorePostgres(db *sql.DB) *SessionStorePostgres {
	return &SessionStorePostgres{db: db}
}
