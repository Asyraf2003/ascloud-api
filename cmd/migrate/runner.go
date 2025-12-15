package main

import (
	"database/sql"
	"fmt"
)

type runner struct {
	db  *sql.DB
	dir string
}

func newRunner(db *sql.DB, dir string) *runner {
	return &runner{db: db, dir: dir}
}

func (r *runner) Up() (int, error) {
	if err := ensureSchemaMigrations(r.db); err != nil {
		return 0, err
	}

	applied, err := loadApplied(r.db)
	if err != nil {
		return 0, err
	}

	migs, err := listMigrations(r.dir)
	if err != nil {
		return 0, err
	}

	appliedN := 0
	for _, m := range migs {
		if applied[m.Name] {
			continue
		}
		if err := applyMigration(r.db, m); err != nil {
			return appliedN, err
		}
		appliedN++
	}
	return appliedN, nil
}

func (r *runner) Status() error {
	if err := ensureSchemaMigrations(r.db); err != nil {
		return err
	}

	applied, err := loadApplied(r.db)
	if err != nil {
		return err
	}

	migs, err := listMigrations(r.dir)
	if err != nil {
		return err
	}

	for _, m := range migs {
		if applied[m.Name] {
			fmt.Printf("[x] %s\n", m.Name)
		} else {
			fmt.Printf("[ ] %s\n", m.Name)
		}
	}
	return nil
}
