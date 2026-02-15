# ADR 0015: Step 3 — Default API Runner + Auth Core Storage on DynamoDB (AWS-first)

Tanggal: 2026-02-15  
Status: Accepted

## Context (Keadaan awal)
- Default runtime `cmd/api` sebelumnya `panic(...)` kecuali dijalankan dengan `-tags=legacy_postgres`.
- Implementasi core auth (session, identity, account, audit sink) tersedia dalam `internal/modules/auth/store/postgres` dan di-wire lewat `internal/modules/auth/wire/legacy_google_wire_postgres.go` (build tag `legacy_postgres`).
- MVP hosting statis butuh fondasi data inti untuk dipakai Step 4 (upload pipeline event-driven): account/identity/session/audit minimal.
- Repo menerapkan boundary hexagonal (usecase tidak boleh tergantung vendor) dan kontrak transport stabil; error handling wajib via `presenter.HTTPErrorHandler`.
- Produk fokus AWS (region Jakarta: `ap-southeast-3`).

## Problem (Masalah)
- Runtime default tidak bisa dipakai untuk development/CI tanpa Postgres legacy.
- Fondasi data inti untuk auth/account/session/audit belum ada pada storage utama yang disepakati untuk MVP (DynamoDB).
- Harus memastikan:
  - port contracts tetap bersih,
  - wiring default berjalan (tanpa `panic`),
  - jalur auth tetap sesuai contract existing (`/v1/auth/*`),
  - workflow dev tidak memaksa penggunaan root AWS.

## Options (Opsi)
1) Tetap Postgres legacy sebagai default
- Pro: cepat (sudah ada).
- Kontra: bertentangan dengan keputusan MVP (DynamoDB), membuat jalur AWS-first tertunda.

2) DynamoDB Local untuk dev (docker `amazon/dynamodb-local`) + script create table
- Pro: offline-friendly, murah, cepat untuk iterasi.
- Kontra: ada perbedaan perilaku dibanding AWS real di beberapa edge case; perlu disiplin menyamakan schema/naming.

3) AWS DynamoDB dev (CloudFormation stack) sebagai sumber kebenaran sejak awal
- Pro: paling dekat real-prod; schema & naming terkunci di IaC; melatih workflow AWS Console/CLI sejak dini.
- Kontra: butuh kredensial AWS; tergantung jaringan; ada biaya (kecil untuk dev, tapi tetap ada).

## Decision (Keputusan)
Dipilih **Opsi 3**: **AWS DynamoDB dev via CloudFormation** sebagai default untuk Step 3.

Step 3 menghasilkan:
1) **Default runtime (`cmd/api`) berjalan tanpa `legacy_postgres`**
- Runner default tidak lagi `panic`.
- Runner default menggunakan DynamoDB client (AWS SDK v2) dan tidak bergantung pada Postgres.
- `server.New(log, db any, allowedOrigins)` tetap menerima `db` sebagai `any`; pada default runner nilainya berupa `*dynamodb.Client`.

2) **Auth core persistence menggunakan DynamoDB (multi-table, MVP)**
Port yang dipenuhi (tanpa ubah usecase):
- `ports.AccountService` (create account minimal)
- `ports.IdentityRepository` (find/upsert identity provider=google, sub)
- `ports.SessionStore` (create, get by id, get by refresh hash, rotate + reuse detection, revoke)
- `ports.AuditSink` (record event auth/deploy internal)

Tabel DynamoDB MVP yang dipakai:
- `accounts`
- `auth_identities`
- `sessions`
- `audit_events`

Di environment dev (AWS) tabel diprefix dan diverifikasi ada:
- `your-api-dev-accounts`
- `your-api-dev-auth-identities`
- `your-api-dev-sessions`
- `your-api-dev-audit-events`

Sumber kebenaran provisioning tabel: `deploy/dev/ddb.yml`.

3) **Desain `sessions` tanpa GSI (index-item pattern)**
Untuk memenuhi port `SessionStore`, termasuk `GetByRefreshTokenHash` serta single-session + refresh rotation:
- Session item: `PK = sid#<session_id>`
- Refresh hash index item: `PK = rh#<hash>` → menyimpan `sid` dan `kind` (`cur`/`prev`)
- Single-session pointer item: `PK = usr#<user_id>` → menyimpan `current_sid`

Konsekuensi desain:
- `GetByRefreshTokenHash(hash)` dilakukan via lookup `rh#hash` → resolve `sid` → load session.
- Rotation mempromosikan `cur` → `prev` dan menulis index baru untuk hash baru.
- Reuse detection dilakukan dengan menolak kondisi yang tidak valid dan/atau mendeteksi hash yang mengarah ke `prev` (indikasi reuse/invalid).

4) **Wiring auth non-legacy**
- Menambah wiring non-legacy: `WireAuthGoogle(*dynamodb.Client, config.AuthConfig)` (tanpa build tag `legacy_postgres`).
- Handler diset via holder pattern (kontrak transport tetap):
  - `authHTTP.SetGoogleHandler(...)`
  - `authHTTP.SetSessionHandler(...)`
- Kontrak transport auth tidak berubah:
  - `GET /v1/auth/google/start`
  - `GET /v1/auth/google/callback`
  - `POST /v1/auth/refresh`
  - `POST /v1/auth/logout`
- Semua error response tetap melalui `presenter.HTTPErrorHandler` (presenter envelope).

5) **Dev ergonomics (konsistensi env)**
- Dev workflow mengandalkan `.env` untuk local run yang konsisten.
- Kredensial AWS untuk CLI ditangani via `AWS_PROFILE` (IAM user, bukan root), bukan menaruh access key di repo.
- (Catatan operasional) `.env` tidak boleh di-commit; gunakan `.env.example` + `.gitignore`.

## Alternatives (Alternatif yang dipertimbangkan)
- Postgres sebagai default runner: ditolak karena tidak AWS-first dan mempertahankan ketergantungan legacy.
- DynamoDB single-table design: ditunda (kompleks, tidak diperlukan untuk Step 3).
- GSI untuk lookup refresh-hash: ditolak untuk MVP (mengikat schema/index lebih awal; index-item pattern cukup untuk kebutuhan sekarang).

## Consequences (Dampak)
Positif:
- Default runtime tidak membutuhkan Postgres legacy untuk start.
- Usecase tetap bersih (tidak bergantung AWS SDK); vendor logic berada di adapter/store.
- Kontrak transport tidak berubah; router/presenter tetap stabil.
- Dev environment lebih dekat ke real-prod, meminimalkan “works on my machine” saat Step 4 dan seterusnya.
- Workflow AWS lebih aman: memakai IAM user + AWS profile, bukan root.

Negatif/Risiko:
- Dev bergantung pada kredensial AWS dan koneksi jaringan.
- Ada biaya AWS (kecil untuk dev, tapi tetap ada).
- Perlu disiplin operasional:
  - jangan commit `.env`,
  - rotasi access key bila perlu,
  - least-privilege policy dapat ditingkatkan kemudian (awal dev memakai admin untuk kecepatan, bukan posture final).

## Security Notes (Catatan keamanan)
- Refresh token tidak disimpan plain, hanya hash; tetap memakai pepper (`AUTH_REFRESH_PEPPER`).
- Reuse detection + single-session enforcement dilakukan di layer storage.
- Data `Meta` disimpan internal (string JSON) dan tidak dikirim mentah ke client (policy JSONB).

## Verification (DoD Step 3)
- `go test ./...`
- `make audit`
- `go run ./cmd/api` memulai server tanpa `panic`
- `curl -i http://localhost:8080/v1/health` → `200 OK`
- `curl -i 'http://localhost:8080/v1/auth/google/start?purpose=login'` → `302 Found` + `Location: accounts.google.com/...`
- Catatan operasional: error “port already in use” dianggap isu lokal (operasional), bukan kegagalan wiring/runtime.
