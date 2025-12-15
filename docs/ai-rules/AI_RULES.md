# AI_RULES.md

Dokumen ini adalah aturan kerja untuk perubahan kode di repo ini.
Tujuan: perubahan konsisten, scalable, minim efek berantai, dan lolos audit.

---

## Repo Identity
- Module path: [MODULE_PATH]  # contoh: example.com/your-api
- Primary language: Go
- Framework HTTP: Echo
- Default API address: [HTTP_ADDR_DEFAULT]  # contoh: :8080
- Timezone (opsional): [TIMEZONE]  # contoh: Asia/Jakarta

---

## Non-negotiables (Hard Rules)
1) Jangan ubah signature publik ini:
   - `internal/transport/http/router.Register(*echo.Echo)`
   - `internal/transport/http/router/v1.Register(*echo.Echo)` (atau `v1.Register(*echo.Echo)` sesuai implementasi)
   - `internal/transport/http/presenter.HTTPErrorHandler(error, echo.Context)`

2) 1 folder = 1 package (tidak boleh campur package dalam satu folder).

3) File size limit:
   - Tidak boleh ada file > 100 baris.
   - Split berdasarkan tanggung jawab (router per area, presenter per concern, usecase per fitur).
   - Jika harus >100 baris untuk alasan kuat: tulis alasan di komentar `#TODO: justify`.

4) Debug endpoints:
   - Tidak boleh menambah endpoint debug kecuali diminta.
   - Debug routes wajib gated: `DEBUG_ROUTES=1`.

5) Error response:
   - Semua error response untuk user wajib berbentuk JSON envelope melalui `presenter.HTTPErrorHandler`.
   - Dilarang mengirim stacktrace, vendor error mentah, SQL error mentah, token/cookie/secret.

6) JSONB policy:
   - JSONB hanya untuk storage/meta internal (audit/flexible fields).
   - JSONB/meta tidak pernah dikirim mentah ke client.
   - Audit meta wajib allowlist/redact (gunakan `redact.RedactMap` atau allowlist map).

---

## Architecture Boundaries (Anti efek berantai)
- HTTP layer (`internal/transport/http/...`) hanya:
  - parse request
  - validasi ringan
  - panggil usecase
  - return response via presenter
- Usecase layer (`internal/modules/*/usecase`) hanya bergantung pada:
  - domain (`internal/modules/*/domain`)
  - ports (`internal/modules/*/ports`)
  - shared utilities yang netral (contoh `internal/shared/...`)
  - Tidak boleh import `internal/transport/http/*` atau vendor/cloud sdk.
- Vendor/IO adapters hanya di `internal/platform/*` dan implement interfaces di `ports`.
- Error type standar:
  - gunakan `internal/shared/apperr.AppError` untuk error yang keluar dari usecase
  - presenter memetakan `AppError` → response aman

---

## Folder Contracts
### Router
- Router induk: `internal/transport/http/router/router.go` (entrypoint)
- Subrouter:
  - `internal/transport/http/router/health`
  - `internal/transport/http/router/v1`
  - `internal/transport/http/router/v2`
  - `internal/transport/http/router/debug` (gated)
  - `internal/transport/http/router/audit`

Rule: jika menambah endpoint v1, harus masuk ke `internal/transport/http/router/v1/*` (bukan di router induk).

### Presenter
- `internal/transport/http/presenter/error.go` adalah pusat sanitasi error response + redaction log.
- Presenter response dibagi:
  - `success.go`, `auth.go`, `billing.go`, `envelope.go`, `error.go`

Rule: handler tidak boleh membuat format response sendiri yang keluar dari standar presenter.

---

## Security Baseline
- Tidak ada secrets di repo.
- Jangan log Authorization/Cookie/ApiKey:
  - wajib redaction.
- Login/session (jika ada):
  - refresh token tidak disimpan plain (store hash)
  - rotation anti-reuse
  - revoke support
- Rate limiting (minimum plan): [RATE_LIMIT_PLAN] #TODO: define later

---

## Definition of Done (Setiap PR/Task)
Wajib menyertakan:
1) Daftar file baru/berubah.
2) Isi final tiap file (bukan potongan).
3) Perintah terminal:
   - `gofmt -w .`
   - `go test ./...`
   - `go vet ./...`
4) Curl sanity tests + expected output:
   - `GET /health` → 200 JSON
   - `GET /v1/health` → 200 JSON
   - `GET /ga-ada` → 404 JSON error envelope
   - `GET /__debug/ping` → 200 JSON hanya saat `DEBUG_ROUTES=1`

---

## Required Snapshot Before Starting Any Work
Sebelum mengusulkan blueprint, AI harus meminta output:
- `tree -L 6 internal/transport/http`
- `tree -L 6 internal/modules/[TARGET_MODULE]`
- `cat internal/transport/http/router/router.go`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `rg -n "^package " internal/transport/http/router`
- (opsional) file terkait fitur: [EXTRA_SNAPSHOT_FILES]

---

## Roadmap Hook
Jika ada perubahan besar, tulis ADR ringkas di:
- `docs/adr/` (judul + keputusan + tradeoff + konsekuensi)


<!--  -->


Bagian yang WAJIB kamu isi

[MODULE_PATH] (paling penting)

[HTTP_ADDR_DEFAULT]

[TARGET_MODULE] per task (misal auth)

[EXTRA_SNAPSHOT_FILES] kalau ada modul spesifik (misal internal/platform/google/...)

[RATE_LIMIT_PLAN] boleh TBD dulu