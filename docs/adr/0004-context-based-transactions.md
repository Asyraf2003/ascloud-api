# ADR 0004: Context-based Transactions (Transactor + Executor via Context)

Tanggal: 2025-12-30 09:00 WITA  
Status: Accepted

## Context

Beberapa flow autentikasi (contoh: Google OIDC callback) membutuhkan operasi database multi-step yang harus *atomic*:

- Create/resolve `accounts`
- Link `auth_identities` ke account tersebut
- (lalu lanjut issue session/tokens)

Tanpa transaksi, operasi write ini bisa menghasilkan **partial success** (contoh: account berhasil dibuat, tapi identity gagal ter-link), meninggalkan data “yatim” yang merusak integritas sistem dan menyulitkan audit/debug.

Di sisi lain, transaksi tidak boleh “bocor” ke layer usecase dalam bentuk:
- usecase memegang `*sql.Tx`
- usecase melakukan `Commit/Rollback`
- usecase mengimpor layer platform/vendor

Karena itu diperlukan pola transaksi yang:
- memastikan atomicity (ACID: Atomicity minimal)
- tetap menjaga boundary hexagonal (usecase bersih dari vendor/IO)
- mudah dites (unit test tanpa DB, integration test untuk atomicity nyata)

## Decision

Mengadopsi pola **Context-based Transaction** dengan kontrak `Transactor` di layer ports.

### Komponen utama

1) Platform Postgres transaction utilities  
Lokasi: `internal/platform/datastore/postgres/tx.go`

- `WithTx(ctx, *sql.Tx)` menyisipkan transaksi ke `context.Context`
- `GetExecutor(ctx, defaultDB)` mengembalikan executor:
  - `*sql.Tx` jika ada transaksi di context
  - `*sql.DB` jika tidak ada
- `RunInTx(ctx, db, fn)` menjalankan `fn` di dalam transaksi (commit/rollback)

2) Kontrak transaksi pada module auth  
Lokasi: `internal/modules/auth/ports/transactor.go`

- `ports.Transactor` menjadi dependency usecase untuk “unit of work”
- `ports.NoopTransactor` untuk unit test (menjalankan `fn` tanpa DB)

3) Usecase membatasi transaction boundary  
Contoh lokasi: `internal/modules/auth/usecase/google_callback.go`

- usecase membungkus operasi multi-write dalam `u.tx.RunInTx(ctx, fn)`
- trust decision / issue token dapat dilakukan di luar transaksi karena tidak mengubah state DB inti

4) Store/repository Postgres harus TX-aware  
Contoh lokasi:
- `internal/modules/auth/store/postgres/auth_account_service.go`
- `internal/modules/auth/store/postgres/auth_identity_repo.go`

Aturan: setiap query/exec di adapter Postgres wajib melalui:

- `executor := postgres.GetExecutor(ctx, db)`
- lalu `executor.QueryRowContext(...)` / `executor.ExecContext(...)`

Sehingga jika usecase membuka transaksi, seluruh operasi store otomatis menggunakan `*sql.Tx` yang sama.

## Scope Note (AWS-first vs Legacy Postgres)

ADR ini mendefinisikan pola transaksi untuk **adapter Postgres (legacy path)**.

Untuk jalur AWS-first (DynamoDB), atomicity dicapai melalui:
- conditional writes / idempotency keys, atau
- DynamoDB TransactWriteItems
dengan tetap menjaga boundary melalui kontrak ports yang setara (tanpa membawa detail vendor ke usecase).

## Consequences

### Positive
- **Atomicity terjaga**: operasi multi-step “all-or-nothing”, tidak ada data setengah jadi.
- **Boundary tetap bersih**: usecase tidak mengimpor vendor/platform DB.
- **Scalable untuk future flow**: pola yang sama bisa dipakai untuk flow lain (refresh rotation, revoke, audit append, dsb).
- **Testability meningkat**:
  - unit test bisa memakai `NoopTransactor`
  - integration test bisa validasi rollback/commit terhadap Postgres nyata

### Negative / Trade-offs
- Store adapter harus disiplin memakai `GetExecutor(ctx, db)` (kalau lupa, transaksi “bocor” jadi non-atomic).
- Context membawa state transaksi, sehingga perlu konsistensi penggunaan `context.Context` di semua call chain.
- Pola ini spesifik untuk adapter yang mendukung “executor dual mode” (`*sql.DB` vs `*sql.Tx`).
  Untuk teknologi datastore lain, perlu pola ekivalen.

## Alternatives Considered

1) Pass `*sql.Tx` langsung ke usecase/store  
Ditolak: bocor vendor ke business layer, melemahkan hexagonal boundary.

2) Repository mengatur transaksi sendiri  
Ditolak: transaction boundary jadi tersebar, sulit diaudit, mudah terjadi nested/overlapping tx yang tidak jelas.

3) Mengadopsi ORM / library lain (sqlx/pgx native) untuk mengatur unit-of-work  
Ditunda: bisa dilakukan nanti, tetapi tetap harus mempertahankan prinsip boundary. Pola kontrak `Transactor` masih valid bahkan jika implementasi DB berubah.

## Rules (Wajib)

- Usecase tidak boleh import `internal/platform/datastore/postgres`.
- Usecase hanya depend pada `ports.Transactor`.
- Adapter Postgres wajib query/exec via `postgres.GetExecutor(ctx, db)`.
- Atomicity behavior (rollback/commit) wajib diuji via integration test (`//go:build integration`) ketika flow sudah punya skenario error penting.