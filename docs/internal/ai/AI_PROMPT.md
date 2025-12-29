# AI_PROMPT

Template prompt untuk tiap task saat minta AI kerja di repo ini.

Tujuan:
- AI kerja dari **fakta repo** (snapshot), bukan ngarang.
- Keputusan produk jelas.
- Ada **blueprint dulu** sebelum eksekusi.
- Perubahan konsisten, minim efek berantai, dan lolos audit.

> Jika ada konflik, `docs/internal/ai/AI_RULES.md` menang.

---

## Template Prompt (copy-paste)

### REPO CONTEXT
- Module: `[MODULE_PATH]` (harus sama dengan `go.mod`)
- Ikuti: `docs/internal/ai/AI_RULES.md` (hard rules + boundaries + DoD).
- Kontrak stabil (jangan diubah):
  - `internal/transport/http/router.Register(*echo.Echo)`
  - `internal/transport/http/router/v1.Register(*echo.Echo)`
  - `internal/transport/http/presenter.HTTPErrorHandler(error, echo.Context)`
- HTTP stack:
  - Router: `internal/transport/http/router/*`
  - Presenter: `internal/transport/http/presenter/*`
- Error response: JSON envelope via `presenter.HTTPErrorHandler` (no secrets).
- Debug routes gated: `DEBUG_ROUTES=1`.
- 1 folder = 1 package. File <= 100 baris (split by responsibility).
- Scope rule: **jangan refactor di luar task** kecuali perlu untuk fix test/audit.

### TASK
Tulis spesifik (1–3 kalimat), termasuk boundary modulnya.
- Target module: `[TARGET_MODULE]`
- Tujuan: …
- Endpoint/flow yang diubah: …
- Constraint penting: …

Contoh:
- “Tambah endpoint `POST /v1/hosting/projects` untuk create project hosting statis di module `hosting`.”

---

## REQUIRED SNAPSHOT (WAJIB, paste output)
> AI tidak boleh bikin blueprint sebelum snapshot ini ada.

### 1) Struktur
- `tree -L 6 internal/transport/http`
- `tree -L 6 internal/modules/[TARGET_MODULE]`

### 2) Kontrak inti (wajib ada)
- `cat internal/transport/http/router/router.go`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `rg -n "^package " internal/transport/http/router`

### 3) Extra snapshot (kalau relevan)
Tulis daftar file yang wajib dibaca AI untuk task ini:
- `[EXTRA_SNAPSHOT_FILES]`

---

## DECISIONS (jawab kalau belum jelas)
AI wajib berhenti dan tanya kalau keputusan ini mempengaruhi security/contract.

- Target client: web browser / API-to-API / keduanya?
- Token model: JWT access + refresh rotation / session opaque?
- Token delivery: refresh via HttpOnly cookie / via response body?
- Expiry: access (mis. 15m/30m), refresh (mis. 14d/30d)?
- Apakah butuh trust score / step-up (AAL2)?
- Audit events minimal apa yang wajib dicatat?

---

## WORKFLOW (AI wajib ikuti)
1) **Ringkas requirement final** (bullet list).  
   Jika ada info kurang, tulis “MISSING” dan tanya, jangan asumsi diam-diam.

2) **Kritik & tradeoff**  
   Sebut risiko pilihan yang berbahaya + alternatif best practice yang lebih aman.

3) **Blueprint (sebelum eksekusi)**
   Wajib berisi:
   - Endpoint/flow (request → usecase → ports → output)
   - Kontrak ports (interface + ownership modul)
   - Data ownership (siapa punya apa)
   - Daftar file baru/ubah (path jelas)
   - Test plan (unit/component/integration mana)

4) **Eksekusi implementasi hanya setelah blueprint disetujui.**

---

## DELIVERABLES (saat eksekusi)
- Daftar file baru/berubah.
- Isi final tiap file (full file, bukan potongan).
- Command DoD:
  - `gofmt -w .`
  - `go test ./... -count=1`
  - `go vet ./...`
  - `make audit` (kalau ada)
- Curl sanity (kalau HTTP) + expected response.
- Jika ada bug: fix minimal dan tetap patuhi boundary.

---

## Quick Header untuk Chat Baru (ringkas)

REPO HEADER
- Module: [MODULE_PATH]
- Ikuti `docs/internal/ai/AI_RULES.md` (hard rules).
- Kontrak stabil: `router.Register`, `v1.Register`, `presenter.HTTPErrorHandler`.
- Error response: JSON envelope (no secrets).
- Debug gated: `DEBUG_ROUTES=1`.
- 1 folder = 1 package. File <=100 baris.
- DoD: gofmt + test + vet + make audit + sanity curl.

SNAPSHOT WAJIB (paste output):
- `tree -L 6 internal/transport/http`
- `tree -L 6 internal/modules/[TARGET_MODULE]`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `rg -n "^package " internal/transport/http/router`
- [EXTRA_SNAPSHOT_FILES]
