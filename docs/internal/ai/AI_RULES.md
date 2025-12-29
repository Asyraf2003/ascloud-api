# AI_RULES

Dokumen ini adalah aturan kerja saat AI membantu perubahan di repo ini.
Tujuan: perubahan konsisten, minim efek berantai, gampang diaudit, dan aman untuk security.

Jika ada konflik antara “preferensi gaya” vs aturan di sini: **aturan di sini menang**.

---

## Repo Identity (konstanta)
- Module path: `example.com/your-api`
- Language: Go
- HTTP framework: Echo
- Default address: `:8080`
- Timezone: Asia/Makassar (WITA)

---

## Non-negotiables (Hard Rules)

### 1) Public contracts tidak boleh diubah
Dilarang mengubah signature publik ini:
- `internal/transport/http/router.Register(*echo.Echo)`
- `internal/transport/http/router/v1.Register(*echo.Echo)` (atau signature yang setara sesuai implementasi repo)
- `internal/transport/http/presenter.HTTPErrorHandler(error, echo.Context)`

Jika memang butuh perubahan kontrak:
- buat ADR dulu,
- update dokumen,
- baru ubah kode (dan siap migrasi dampaknya).

### 2) 1 folder = 1 package
Tidak boleh ada lebih dari satu `package` dalam satu folder (non-test).

### 3) File size limit
- Target: file Go ≤ 100 baris.
- Kalau ada file > 100 baris:
  - wajib split berdasar tanggung jawab, atau
  - tulis alasan kuat dengan komentar `// TODO: justify >100 lines` dan rencana split.

### 4) Debug endpoints wajib gated
- Jangan menambah endpoint debug kecuali diminta.
- Debug routes **wajib** gated: `DEBUG_ROUTES=1`.

### 5) Error response wajib lewat presenter
- Semua error ke client wajib JSON envelope via `presenter.HTTPErrorHandler`.
- Dilarang mengirim:
  - stacktrace
  - SQL/vendor error mentah
  - token/cookie/secret
  - payload sensitif mentah

### 6) Secrets & logging policy
- Tidak ada secrets di repo.
- Jangan log Authorization/Cookie/ApiKey/token:
  - wajib redact/mask sebelum log.
- Log idealnya terstruktur + include `request_id`.

---

## Architecture Boundaries (Anti efek berantai)

> Enforcement utama: `scripts/audit_boundaries.sh` + `make audit`.

### Domain (`internal/modules/*/domain`)
Boleh import:
- standard library
- `internal/shared/*` yang **pure** (sesuai ADR 0003)

Dilarang import:
- `internal/platform/*`
- `internal/transport/http/*`
- module lain (`internal/modules/*` selain dirinya)
- third-party packages

### Ports (`internal/modules/*/ports`)
Berisi interface dependency.
Boleh import:
- stdlib
- domain (modul sendiri)
- `internal/shared/*` (pure)

Catatan:
- third-party types boleh kalau benar-benar perlu dan kecil (default: hindari).

### Usecase (`internal/modules/*/usecase`)
Orchestrator flow: validasi ringan, call ports, bentuk output, return error standar.
Boleh import:
- domain + ports + `internal/shared/*` (pure)

Dilarang import:
- `internal/transport/http/*`
- `internal/platform/*`

### Module HTTP Transport (`internal/modules/*/transport/http`)
Tugas:
- parse request
- validasi ringan
- panggil usecase
- return via presenter

Dilarang:
- import `internal/platform/*` (akses IO harus via usecase → ports)

### Core HTTP (`internal/transport/http/...`)
Router/middleware/presenter lintas modul.
Dilarang:
- import `internal/platform/*`

### Platform (`internal/platform/...`)
Implementasi nyata untuk ports (DB, queue, objectstore, edge, token, dll).
Dilarang import:
- `internal/transport/http/*`
- `internal/modules/*/transport/http/*`
- `internal/app/*` (bootstrap/wiring)

Prefer:
- platform expose constructor + implement interface ports.

---

## Security baseline (minimal)
Jika menyentuh auth/session:
- refresh token **tidak** disimpan plain (store hash)
- rotation + reuse detection + revoke
- sesi/identity harus scoped tenant/account (anti IDOR)

Rate limit:
- boleh TBD, tapi jangan membuat endpoint sensitif tanpa rencana minimal (misal: per-IP + per-account).

---

## Definition of Done (setiap task/PR)
Wajib ada:
1) Daftar file baru/berubah.
2) Isi final tiap file (bukan potongan).
3) Command:
   - `gofmt -w .`
   - `go test ./... -count=1`
   - `go vet ./...`
4) Sanity check (minimal):
   - `GET /health` → 200 JSON
   - `GET /v1/health` → 200 JSON
   - `GET /ga-ada` → 404 JSON error envelope
   - `GET /__debug/ping` → 200 JSON hanya saat `DEBUG_ROUTES=1`

---

## Required Snapshot (sebelum AI mulai kerja)
AI wajib minta output ini dulu, jangan asumsi:
- `tree -L 6 internal/transport/http`
- `tree -L 6 internal/modules/<TARGET>`
- `cat internal/transport/http/router/router.go`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `rg -n "^package " internal/transport/http/router`
- tambahan file relevan bila task menyentuh area spesifik (middleware/platform/schema)
