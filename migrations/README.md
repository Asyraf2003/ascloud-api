# Migrations (Postgres)

Jalur resmi:
- Apply migrations: `go run ./cmd/migrate up`
- Lihat status: `go run ./cmd/migrate status`

Env yang dibaca:
- `DB_DSN` (utama)
- fallback: `DATABASE_URL`
