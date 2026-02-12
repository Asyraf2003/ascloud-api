//go:build !legacy_postgres
// +build !legacy_postgres

package main

import "fmt"

func main() {
	fmt.Println("cmd/migrate is legacy Postgres tooling. Run with: go run -tags=legacy_postgres ./cmd/migrate <cmd>")
}
