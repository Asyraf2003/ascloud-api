//go:build legacy_postgres
// +build legacy_postgres

package postgres

import (
	"context"
	"database/sql"
)

// SQLQueryer adalah interface agar repo bisa menerima *sql.DB atau *sql.Tx
type SQLQueryer interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type txKey struct{}

// WithTx memasukkan *sql.Tx ke dalam context
func WithTx(ctx context.Context, tx *sql.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// GetExecutor mengambil *sql.Tx jika ada di context, jika tidak balikkan *sql.DB
func GetExecutor(ctx context.Context, defaultDB *sql.DB) SQLQueryer {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return tx
	}
	return defaultDB
}

// RunInTx membungkus fungsi 'fn' dalam satu database transaction
func RunInTx(ctx context.Context, db *sql.DB, fn func(context.Context) error) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	// Oper context yang sudah disisipi transaksi ke dalam fungsi
	err = fn(WithTx(ctx, tx))
	if err != nil {
		// Jika fungsi error, batalkan semua perubahan di DB
		_ = tx.Rollback()
		return err
	}

	// Jika sukses, simpan perubahan secara permanen
	return tx.Commit()
}

type Transactor struct{ db *sql.DB }

func NewTransactor(db *sql.DB) *Transactor { return &Transactor{db: db} }

func (t *Transactor) RunInTx(ctx context.Context, fn func(context.Context) error) error {
	return RunInTx(ctx, t.db, fn)
}
