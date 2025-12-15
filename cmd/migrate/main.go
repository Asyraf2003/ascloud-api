package main

import (
	"fmt"
	"os"

	"example.com/your-api/internal/platform/datastore/postgres"
)

func main() {
	cmd := "up"
	if len(os.Args) >= 2 {
		cmd = os.Args[1]
	}

	dsn := getenv("DB_DSN", getenv("DATABASE_URL", ""))
	if dsn == "" {
		fatalf("DB_DSN/DATABASE_URL kosong")
	}

	db, err := postgres.New(dsn)
	if err != nil {
		fatalf("open db: %v", err)
	}
	defer db.Close()

	r := newRunner(db, "migrations")

	switch cmd {
	case "up":
		n, err := r.Up()
		if err != nil {
			fatalf("migrate up: %v", err)
		}
		fmt.Printf("ok: applied %d migration(s)\n", n)
	case "status":
		if err := r.Status(); err != nil {
			fatalf("status: %v", err)
		}
	default:
		fatalf("unknown cmd=%s (use: up|status)", cmd)
	}
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func fatalf(f string, a ...any) {
	fmt.Fprintf(os.Stderr, f+"\n", a...)
	os.Exit(1)
}
