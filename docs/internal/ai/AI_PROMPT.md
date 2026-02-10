# AI_PROMPT

Template prompt untuk tiap task saat minta AI bantu kerja di repo ini.

Tujuan:
AI kerja dari fakta repo (snapshot), keputusan produk jelas (dengan opsi), lalu blueprint rapi sebelum eksekusi.

---

## Template Prompt (copy-paste)

### REPO CONTEXT
- Module: (paste dari `go.mod`, baris `module ...`)
- Ikuti `docs/internal/ai/AI_RULES.md` (hard rules + boundaries + DoD).
- Router: `internal/transport/http/router/*` (router induk + v1 modular).
- Presenter: `internal/transport/http/presenter/*`.
- Error response: JSON envelope via `presenter.HTTPErrorHandler` (no secrets).
- Debug routes wajib gated: `DEBUG_ROUTES=1`.
- 1 folder = 1 package. File <=100 baris (exception boleh untuk file kritikal dengan `#TODO: justify`).
- Baseline aktif: AWS (CloudFront + S3 + Lambda + SQS + DynamoDB). Provider lain: INACTIVE.

### TASK
- [TULIS TASK DI SINI]
  Contoh: “Tambah endpoint create project hosting statis di module hosting.”

### REQUIRED SNAPSHOT (WAJIB, paste output)
1) Identitas repo:
- `cat go.mod`
2) Struktur:
- `tree -L 6 internal/transport/http`
- `tree -L 6 internal/modules/[TARGET_MODULE]`
3) Kontrak inti:
- `cat internal/transport/http/router/router.go`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `rg -n "^package " internal/transport/http/router`
4) Extra snapshot (kalau relevan):
- [EXTRA_SNAPSHOT_FILES]

### DECISIONS (kalau belum terkunci)
Jika ada keputusan arsitektur/kontrak yang mempengaruhi implementasi, jawab dengan memilih opsi.
Format jawaban AI WAJIB seperti ini:
- Opsi A: ... (plus/minus)
- Opsi B: ... (plus/minus)
- Rekomendasi: ... (alasan sesuai baseline AWS + standar enterprise + blueprint)

Contoh keputusan yang sering muncul:
- DB: DynamoDB vs Postgres
- Edge mapping: CloudFront Function+KVS vs alternatif MVP
- Quota: hard-fail vs soft-suspend + grace
- Auth browser: cookie (same eTLD+1) vs bearer token

### WORKFLOW (AI wajib ikuti)
1) Ringkas requirement final (bullet list).
2) Kritik pilihan yang berbahaya + tawarkan alternatif best practice.
3) Buat blueprint:
   - flow endpoint
   - kontrak ports
   - file list & ownership
   - catatan boundary & risiko
4) Baru eksekusi implementasi setelah blueprint “oke”.

### DELIVERABLES (saat eksekusi)
- Daftar file baru/berubah.
- Isi final tiap file (bukan potongan).
- Command:
  - `gofmt -w .`
  - `go test ./... -count=1`
  - `go vet ./...`
  - `make audit` (kalau tersedia di repo)
- Curl sanity (kalau HTTP) + expected response.
- Jika ada bug, fix hanya bagian terkait (minim efek berantai).

---

## Quick Header untuk Chat Baru (ringkas)

REPO HEADER
- Module: (paste dari `go.mod`)
- Ikuti `docs/internal/ai/AI_RULES.md`.
- Kontrak stabil: `router.Register`, `v1.Register`, `presenter.HTTPErrorHandler`.
- Error response: JSON envelope (no secrets).
- Debug gated: `DEBUG_ROUTES=1`.
- 1 folder = 1 package. File <=100 baris (exception boleh untuk file kritikal dengan `#TODO: justify`).
- Baseline aktif: AWS (CloudFront + S3 + Lambda + SQS + DynamoDB). Provider lain: INACTIVE.
- DoD: gofmt + test + vet + make audit + sanity curl.

SNAPSHOT WAJIB (paste output):
- `cat go.mod`
- `tree -L 6 internal/transport/http`
- `tree -L 6 internal/modules/[TARGET_MODULE]`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `rg -n "^package " internal/transport/http/router`
- [EXTRA_SNAPSHOT_FILES]
