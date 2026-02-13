# ADR 0014: Hardening Auth Cookie Security (CSRF token distribution + deterministic HTTPS)

Tanggal: 2026-02-13
Status: Accepted

## Context
Repo menerapkan model keamanan cookie-based untuk endpoint auth tertentu (minimal: refresh/logout) dengan layering:
- RequireHTTPS (conditional, berdasarkan policy)
- Origin allowlist
- CSRF double-submit (cookie vs header)
- Rate limiting

Arsitektur dan aturan repo:
- Transport layer tipis: parse/validate ringan → usecase → presenter.
- Error response wajib lewat `presenter.HTTPErrorHandler` (JSON envelope).
- Policy auth dibekukan saat bootstrap (InitPolicy) untuk determinisme audit.
- Quality gate wajib: `gofmt`, `go test`, `go vet`, `make audit`.

## Problem
### A) CSRF header membutuhkan token yang tidak punya sumber deterministic
Middleware CSRF (`RequireCSRFWithCode`) mewajibkan:
- Cookie CSRF ada, dan
- Header `X-CSRF-Token` ada, dan nilainya sama persis dengan cookie.

Namun response auth dari:
- `POST /v1/auth/google/callback`
- `POST /v1/auth/refresh`

tidak mengembalikan nilai CSRF token di JSON. Ini membuat client web tidak memiliki sumber token untuk membentuk header CSRF secara deterministik tanpa bergantung pada pembacaan cookie (yang bisa terpengaruh Path/Domain).

### B) HTTPS middleware tidak deterministik saat dipasang
Middleware `trust.RequireHTTPS()` memiliki cabang behavior berdasarkan `APP_ENV` (dev vs non-dev).
Padahal pemasangan middleware HTTPS sudah ditentukan oleh policy router (berdasarkan `authPkg.RequireHTTPS()`), sehingga enforcement HTTPS seharusnya deterministik saat middleware dipasang.

## Options Considered
### 1) CSRF token hanya lewat cookie (JS baca cookie)
- (+) Response lebih kecil.
- (-) Bergantung pada cookie Path/Domain dan perilaku browser.
- (-) Risiko UX dan integrasi client (token header sulit didapat bila cookie tidak terlihat).

### 2) CSRF token dikirim lewat JSON response (dipilih)
- (+) Client selalu punya sumber token untuk membentuk header `X-CSRF-Token`.
- (+) Tetap mempertahankan double-submit check (server tetap bandingkan cookie vs header).
- (+) Perubahan additive (backward compatible).
- (-) CSRF token menjadi data yang di-handle JS client (bukan credential) dan harus dihindari dari logging.

### 3) HTTPS middleware punya mode dev vs prod di dalam middleware
- (+) Nyaman untuk dev.
- (-) Tidak deterministik saat middleware dipasang, bergantung `APP_ENV`.
- (-) Bertabrakan dengan prinsip policy freeze + audit determinism.

### 4) HTTPS enforcement selalu tegas saat middleware dipasang (dipilih)
- (+) Deterministik: jika middleware dipasang, HTTP selalu ditolak.
- (+) Dev HTTP tetap bisa berjalan karena middleware hanya dipasang jika policy mengharuskan.
- (-) Memerlukan test untuk memastikan behavior tidak berubah diam-diam.

## Decision
1) Menambahkan field `csrf_token` pada payload auth (AuthTokens) dan mengembalikannya pada:
   - `POST /v1/auth/google/callback`
   - `POST /v1/auth/refresh`

2) Menjadikan `trust.RequireHTTPS()` deterministik:
   - Jika request bukan HTTPS (berdasarkan `isHTTPS()`), selalu reject 403.
   - Tidak ada pengecualian berbasis `APP_ENV` di dalam middleware.
   - Kebijakan “dev boleh HTTP” ditentukan oleh policy attach (router) melalui `authPkg.RequireHTTPS()`.

## Implementation Notes
Perubahan utama:
- `internal/transport/http/presenter/*`: `AuthTokens` menambahkan `csrf_token`.
- `internal/modules/auth/transport/http/google_handler.go`: menambahkan `CSRFToken` pada response.
- `internal/modules/auth/transport/http/session_handler.go`: menambahkan `CSRFToken` pada response.
- `internal/transport/http/middleware/trust/https_only.go`: enforce HTTPS tanpa logika `APP_ENV`.
- Menambah/menyesuaikan component tests untuk membuktikan:
  - `csrf_token` hadir pada response auth callback & refresh.
  - `RequireHTTPS` menolak HTTP dan menerima request dengan indikator HTTPS.

## Consequences
Positif:
- Client web dapat membentuk header `X-CSRF-Token` secara deterministik.
- Double-submit CSRF tetap kuat (cookie vs header compare).
- HTTPS enforcement menjadi audit-friendly dan tidak bergantung pada `APP_ENV` saat middleware dipasang.

Negatif / Trade-offs:
- CSRF token ikut muncul di response JSON (bukan secret/credential, namun tetap hindari logging).
- `isHTTPS()` masih bergantung pada sinyal TLS/forwarded headers; deployment harus memastikan proxy tepercaya yang mengatur/override header dan tidak mengekspos direct access tanpa sanitasi.

## Verification
Perubahan diverifikasi dengan:
- `gofmt -w internal/transport/http/presenter internal/modules/auth/transport/http internal/transport/http/middleware/trust`
- `go test -tags=component ./internal/modules/auth/transport/http/... -count=1`
- `go test -tags=component ./internal/transport/http/... -count=1`
- `make audit` (unit + component + vet + boundary/package checks + docs/testtags/content audit) PASS.
