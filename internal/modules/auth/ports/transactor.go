package ports

import "context"

// Transactor menjalankan fn dalam 1 transaksi DB.
// Implementasi wajib menyisipkan transaksi ke context (agar repo bisa pakai GetExecutor).
type Transactor interface {
	RunInTx(ctx context.Context, fn func(ctx context.Context) error) error
}

// NoopTransactor untuk unit-test / in-memory store.
type NoopTransactor struct{}

func (NoopTransactor) RunInTx(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
