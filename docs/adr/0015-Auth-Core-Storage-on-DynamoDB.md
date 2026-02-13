# ADR 0015: Step 3 — Default API Runner + Auth Core Storage on DynamoDB (AWS-first)

Date: 2026-02-13  
Status: Accepted

## Context
- Default `cmd/api` sebelumnya `panic(...)` kecuali dijalankan dengan `-tags=legacy_postgres`.
- MVP hosting statis butuh fondasi data inti untuk Step 4 (upload pipeline): account/identity/session/audit minimal.
- Repo menerapkan kontrak transport stabil dan boundary hexagonal; error handling via `presenter.HTTPErrorHandler`.
- Produk fokus AWS (Jakarta: ap-southeast-3).

## Decision
1) **Default runtime (`cmd/api`) berjalan tanpa `legacy_postgres`**
- Runner default menggunakan DynamoDB client (AWS SDK v2) dan tidak bergantung pada Postgres.
- `server.New(log, db any, allowedOrigins)` menerima `db` sebagai `*dynamodb.Client`.

2) **Auth core persistence menggunakan DynamoDB (multi-table)**
Tabel MVP:
- `accounts`
- `auth_identities`
- `sessions`
- `audit_events`

3) **Desain `sessions` tanpa GSI (index-item pattern)**
Untuk memenuhi port `SessionStore` termasuk `GetByRefreshTokenHash` serta single-session + rotation:
- Session item: `pk = sid#<session_id>`
- Refresh hash index item: `pk = rh#<hash>` → menyimpan `sid` dan `kind` (`cur`/`prev`)
- Single-session pointer item: `pk = usr#<user_id>` → `current_sid`

Konsekuensi:
- `GetByRefreshTokenHash(hash)` dilakukan via lookup `rh#hash` → resolve `sid` → load session.
- Rotation mengubah `cur` menjadi `prev` dan menulis index baru untuk hash baru.
- Reuse detection dilakukan dengan menolak kondisi yang tidak valid dan/atau mendeteksi hash yang mengarah ke `prev`.

4) **Wiring auth non-legacy**
- Menambah `WireAuthGoogle(*dynamodb.Client, config.AuthConfig)` (tanpa build tag legacy).
- Handler di-set via `authHTTP.SetGoogleHandler(...)` dan `authHTTP.SetSessionHandler(...)` (kontrak transport tetap).

5) **Dev ergonomics**
- Menambah `make run-api` yang memuat `.env` secara konsisten (menghindari “works in one terminal tab only” problem).

## Alternatives Considered
- Tetap Postgres untuk default runner (ditolak: tidak AWS-first, tetap bergantung legacy).
- DynamoDB single-table design (ditunda: kompleks, tidak perlu untuk Step 3).
- GSI untuk refresh-hash lookup (ditolak untuk MVP: mengikat schema/index lebih awal; index-item cukup).

## Consequences
- Default API tidak membutuhkan DB_DSN/Postgres untuk start.
- Environment harus menyediakan konfigurasi Auth Google dan AWS/DynamoDB.
- DynamoDB tables harus diprovision untuk dev/prod sebelum flow auth dipakai end-to-end.

## Security Notes
- Refresh token tidak disimpan plain, hanya hash; tetap memakai pepper (`AUTH_REFRESH_PEPPER`).
- Reuse detection + single-session enforcement dilakukan di layer storage.
- Data `Meta` disimpan internal (JSON string) dan tidak dikirim mentah ke client (policy JSONB).

## Verification
- `go test ./...`
- `make audit`
- `make run-api` memulai server tanpa `panic` (port conflict dianggap operasional lokal, bukan kegagalan wiring)
