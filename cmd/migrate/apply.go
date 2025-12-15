package main

import (
	"context"
	"database/sql"
	"os"
	"time"
)

func applyMigration(db *sql.DB, m migrationFile) error {
	b, err := os.ReadFile(m.Path)
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	tx, err := db.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelSerializable})
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, string(b)); err != nil {
		return err
	}
	if err := markApplied(tx, m.Name); err != nil {
		return err
	}
	return tx.Commit()
}
