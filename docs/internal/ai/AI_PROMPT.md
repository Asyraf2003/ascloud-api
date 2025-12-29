# AI_PROMPT

Template prompt untuk tiap task saat minta AI bantu kerja di repo ini.

Tujuan:
- AI kerja dari fakta repo (snapshot), bukan ngarang.
- Keputusan produk jelas dulu.
- Blueprint rapi dulu sebelum eksekusi.
- Output siap audit (DoD jelas, minim efek berantai).

---

## Template Prompt (copy-paste)

### REPO CONTEXT
- Module: `[MODULE_PATH]`
- Ikuti `docs/internal/ai/AI_RULES.md` (hard rules + boundaries + DoD).
- Router: `internal/transport/http/router/*` (root + v1 modular).
- Presenter: `internal/transport/http/presenter/*`.
- Error response: JSON envelope via `presenter.HTTPErrorHandler` (no secrets).
- Debug routes wajib gated: `DEBUG_ROUTES=1`.
- 1 folder = 1 package. File <= 100 baris (split by responsibility).

### TASK
- [TULIS TASK DI SINI]
  Contoh: “Tambah endpoint create project hosting statis di module hosting.”

### REQUIRED SNAPSHOT (WAJIB, paste output)
1) Struktur:
- `tree -L 6 internal/transport/http`
- `tree -L 6 internal/modules/[TARGET_MODULE]`

2) Kontrak inti:
- `cat internal/transport/http/router/router.go`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `rg -n "^package " internal/transport/http/router`

3) Extra snapshot (kalau relevan):
- [EXTRA_SNAPSHOT_FILES]

### DECISIONS (jawab kalau belum jelas)
- Target client: web browser / API-to-API / keduanya?
- Token model: JWT access + refresh rotation atau opaque session?
- Token delivery: refresh via HttpOnly cookie atau via response body?
- Expiry: access (mis. 15m/30m), refresh (mis. 14d/30d)?
- Apakah butuh trust score / step-up (AAL2)?
- Audit events minimal apa yang wajib dicatat?

### WORKFLOW (AI wajib ikuti)
1) Ringkas requirement final (bullet list).
2) Kritik pilihan yang berbahaya + tawarkan alternatif best practice.
3) Buat blueprint:
   - flow endpoint
   - kontrak ports
   - file list & ownership (siapa “pemilik data”)
   - test plan (unit/component/integration bila relevan)
4) Baru eksekusi implementasi setelah blueprint disetujui.

### DELIVERABLES (saat eksekusi)
- Daftar file baru/berubah.
- Isi final tiap file (full file, bukan potongan).
- Commands:
  - `gofmt -w .`
  - `go test ./... -count=1`
  - `go vet ./...`
  - `make audit` (kalau tersedia)
- Sanity (kalau HTTP) + expected response.

---

## Quick Header untuk Chat Baru (ringkas)

REPO HEADER
- Module: `[MODULE_PATH]`
- Ikuti `docs/internal/ai/AI_RULES.md`.
- Kontrak stabil: `router.Register`, `v1.Register`, `presenter.HTTPErrorHandler`.
- Error response: JSON envelope (no secrets).
- Debug gated: `DEBUG_ROUTES=1`.
- 1 folder = 1 package. File <= 100 baris.
- DoD: gofmt + test + vet + sanity curl.

SNAPSHOT WAJIB (paste output):
- `tree -L 6 internal/transport/http`
- `tree -L 6 internal/modules/[TARGET_MODULE]`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `rg -n "^package " internal/transport/http/router`
- [EXTRA_SNAPSHOT_FILES]
