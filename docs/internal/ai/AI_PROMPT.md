# AI_PROMPT

Template prompt untuk meminta AI mengerjakan task di repo ini.
Tujuan: AI tidak mengarang, tidak bikin efek berantai, dan hasilnya bisa kamu audit cepat.

> Aturan utama ada di `docs/internal/ai/AI_RULES.md`. Kalau ada konflik, AI_RULES menang.

---

## 1) Prompt Template (copy-paste)

### REPO CONTEXT
- Module: `example.com/your-api`
- Ikuti `docs/internal/ai/AI_RULES.md` (hard rules + boundaries + DoD).
- Router core: `internal/transport/http/router/*` (router induk + v1 modular).
- Presenter: `internal/transport/http/presenter/*`.
- Error response: JSON envelope via `presenter.HTTPErrorHandler` (tanpa secrets).
- Debug routes gated: `DEBUG_ROUTES=1`.
- 1 folder = 1 package. Target file Go ≤ 100 baris.

### TASK
- [TULIS TASK DI SINI]
  (jelaskan tujuan, scope, dan constraint penting)

### REQUIRED SNAPSHOT (WAJIB, jangan asumsi)
Tempel output command ini sebelum AI bikin blueprint:
- `tree -L 6 internal/transport/http`
- `tree -L 6 internal/modules/[TARGET_MODULE]`
- `cat internal/transport/http/router/router.go`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `rg -n "^package " internal/transport/http/router`
- [EXTRA SNAPSHOT FILES] (opsional, bila task menyentuh area spesifik)

### DECISIONS (WAJIB bila belum jelas)
Jawab poin yang relevan:
- Target client: web browser / API-to-API / keduanya?
- Token delivery: refresh via HttpOnly cookie atau via response body?
- Expiry access token & refresh token?
- Apakah butuh step-up (AAL2) / trust score?
- Audit events minimal apa saja yang wajib dicatat?
- Dependency real apa yang boleh dipakai untuk integration test?

### WORKFLOW (urutan kerja AI)
Setelah snapshot + keputusan lengkap:
1) Ringkas requirement final (bullet list).
2) Kritik risiko dan tawarkan alternatif best practice (kalau ada).
3) Buat blueprint (flow + boundaries + file list).
4) Setelah blueprint jelas, lanjut implementasi (minimal perubahan, tidak menjalar).
5) Sediakan DoD: gofmt/test/vet + sanity curl + expected output.

### DELIVERABLES (saat implementasi)
- Daftar file baru/berubah.
- Isi final tiap file (bukan potongan).
- Command:
  - `gofmt -w .`
  - `go test ./... -count=1`
  - `go vet ./...`
- Curl sanity tests + expected response.
- Jika ada bug, perbaiki hanya di file terkait (minim efek berantai).

---

## 2) Header Ringkas (untuk chat baru, copy-paste)
REPO HEADER (ringkas)
- Module: `example.com/your-api`
- Ikuti `docs/internal/ai/AI_RULES.md` (hard rules).
- Kontrak stabil: router.Register, v1.Register, presenter.HTTPErrorHandler.
- Error response: JSON envelope (no secrets).
- Debug routes gated: `DEBUG_ROUTES=1`.
- 1 folder = 1 package. Target file Go ≤ 100 baris.
- DoD: gofmt + go test + go vet + sanity curl.

SNAPSHOT WAJIB (tempel output):
- `tree -L 6 internal/transport/http`
- `tree -L 6 internal/modules/[TARGET]`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `rg -n "^package " internal/transport/http/router`
