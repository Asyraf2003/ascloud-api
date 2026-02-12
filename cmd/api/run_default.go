//go:build !legacy_postgres
// +build !legacy_postgres

package main

func run() {
	panic("cmd/api legacy Postgres disabled by default: build/run with -tags=legacy_postgres")
}
