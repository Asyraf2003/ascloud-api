# AI_RULES

Aturan kerja saat AI membantu perubahan di repo ini.

Tujuan:
- Perubahan konsisten, scalable, minim efek berantai.
- Boundary jelas (hexagonal/modular).
- Output bisa diaudit (DoD jelas).
- AI tidak ngarang (wajib snapshot).

Jika ada konflik dokumen, file ini menang.

---

## Repo Identity
- Module path: `example.com/your-api`
- Primary language: Go
- HTTP framework: Echo
- Default HTTP addr: `:8080`
- Timezone: WITA (`Asia/Makassar`)

---

## Non-negotiables (Hard Rules)

### 1) Protected Public Contracts (jangan diubah)
- `internal/transport/http/router.Register(*echo.Echo)`
- `internal/transport/http/router/v1.Register(*echo.Echo)`
- `internal/transport/http/presenter.HTTPErrorHandler(error, echo.Context)`

### 2) Package & file hygiene
- 1 folder = 1 package (jangan campur package dalam satu folder).
- File tidak boleh > 100 baris.
  - Split by responsibility (router per area, presenter per concern, usecase per flow).
  - Jika ada alasan kuat >100 baris, tulis alasan singkat dengan `#TODO: justify`.

### 3) Debug endpoints
- Dilarang menambah endpoint debug kecuali diminta.
- Debug routes wajib gated: `DEBUG_ROUTES=1`.

### 4) Error response policy (ke user)
- Semua error response untuk user wajib JSON envelope via `presenter.HTTPErrorHandler`.
- Dilarang kirim:
  - stacktrace
  - vendor error mentah
  - SQL error mentah
  - token/cookie/secret
- Semua logging wajib redact: Authorization, Cookie, ApiKey, token apapun.

### 5) JSONB policy
- JSONB hanya untuk storage/meta internal (audit/flexible fields).
- JSONB/meta tidak pernah dikirim mentah ke client.
- Audit meta wajib allowlist/redact.

---

## Architecture Boundaries (Anti efek berantai)

### Domain (`internal/modules/*/domain`)
- Fokus: invariants, entity/value, semantic errors.
- Boleh import:
  - standard library
  - `internal/shared/*` yang pure (lihat ADR 0003)
- Dilarang import:
  - `internal/platform/*`
  - `internal/transport/http/*`
  - module lain (`internal/modules/*`)
  - third-party packages

### Ports (`internal/modules/*/ports`)
- Berisi interface dependency.
- Boleh import:
  - stdlib
  - domain (modul sendiri)
  - `internal/shared/*` (pure)
- Third-party hanya kalau benar-benar perlu dan disetujui (default: minimalkan).

### Usecase (`internal/modules/*/usecase`)
- Orchestrator flow: validasi ringan, call ports, bentuk output, return error terstandar.
- Boleh import: domain + ports + `internal/shared/*` (pure).
- Dilarang import:
  - `internal/transport/http/*`
  - `internal/platform/*`
  - vendor/cloud sdk

### Module HTTP transport (`internal/modules/*/transport/http`)
- Tugas: mapping request/response + validasi ringan + panggil usecase.
- Dilarang import `internal/platform/*`.
- Response wajib lewat presenter/envelope (jangan bikin format sendiri).

### Core HTTP (`internal/transport/http/...`)
- Router/middleware/presenter lintas modul.
- Dilarang import `internal/platform/*`.

### Platform (`internal/platform/...`)
- Implementasi nyata (DB, queue, objectstore, edge, token, provider).
- Dilarang import:
  - `internal/transport/http/*`
  - `internal/modules/*/transport/http/*`
  - `internal/app/*` (bootstrap/wiring)

Enforcement:
- `scripts/audit_boundaries.sh`
- `make audit`

---

## Folder Contracts

### Router
- Root: `internal/transport/http/router/router.go`
- v1: `internal/transport/http/router/v1/*`
Rule:
- Endpoint v1 harus didaftarkan di `internal/transport/http/router/v1/<area>/routes.go`

### Presenter
- Pusat sanitasi error: `internal/transport/http/presenter/error.go`
Rule:
- Handler tidak boleh bikin response format sendiri di luar presenter.

---

## Security Baseline (minimum)
- Tidak ada secrets di repo.
- Refresh token:
  - tidak disimpan plain (store hash)
  - rotation anti-reuse
  - revoke support
- Rate limiting plan: `TBD` (harus ditetapkan sebelum public launch)

---

## Definition of Done (setiap task/PR)
Wajib menyertakan:
1) Daftar file baru/berubah.
2) Isi final tiap file (full file, bukan potongan).
3) Commands:
   - `gofmt -w .`
   - `go test ./... -count=1`
   - `go vet ./...`
   - `make audit` (kalau tersedia)
4) Sanity (kalau HTTP):
   - `GET /health` → 200 JSON
   - `GET /v1/health` → 200 JSON
   - `GET /ga-ada` → 404 JSON error envelope
   - `GET /__debug/ping` → 200 JSON hanya saat `DEBUG_ROUTES=1`

---

## Required Snapshot (sebelum blueprint)
AI wajib minta dan user wajib paste output:
- `tree -L 6 internal/transport/http`
- `tree -L 6 internal/modules/<target>`
- `cat internal/transport/http/router/router.go`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `rg -n "^package " internal/transport/http/router`
- (extra) file terkait fitur bila task menyentuh area spesifik

---

## Update Policy (biar tidak membusuk)
- Kontrak/boundary berubah → update file ini + buat ADR bila relevan.
- Struktur/path berubah → update snapshot list (jangan ada path ngaco).
