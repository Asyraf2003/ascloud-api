# EXAMPLE_PROMPT (Auth Google)

Contoh prompt yang sudah terisi. Tujuan: jadi referensi bentuk final yang benar, bukan sekadar teori.

> Catatan penting: contoh ini tidak boleh menyebut path file yang “ngarang”.
> Kalau ragu nama file, pakai `tree`/`find` di snapshot, atau refer ke folder-nya saja.

---

## REPO CONTEXT
- Module: `example.com/your-api`
- Ikuti `docs/internal/ai/AI_RULES.md` (hard rules + boundaries + DoD).
- Router: `internal/transport/http/router/*`
- Presenter: `internal/transport/http/presenter/*`
- Error response: JSON envelope via `presenter.HTTPErrorHandler` (no secrets).
- Debug routes gated: `DEBUG_ROUTES=1`
- 1 folder = 1 package. File <= 100 baris.

---

## TASK
Implement login/register via **Google OIDC** dengan:
- Access token: **JWT** (exp 30m)
- Refresh token: **HttpOnly cookie**
- Refresh rotation + reuse detection
- Single-session (login baru revoke session lama)
- Step-up auth (AAL2)
- Audit events untuk seluruh aksi auth penting

Endpoint (contoh, sesuaikan repo):
- `POST /v1/auth/google/start`
- `GET  /v1/auth/google/callback`
- `POST /v1/auth/refresh`
- `POST /v1/auth/logout`
- `POST /v1/auth/google/stepup/start`
- `GET  /v1/auth/google/stepup/callback`

---

## REQUIRED SNAPSHOT (WAJIB, paste output)
1) Struktur & kontrak HTTP:
- `tree -L 6 internal/transport/http`
- `cat internal/transport/http/router/router.go`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `rg -n "^package " internal/transport/http/router`

2) Modul auth:
- `tree -L 6 internal/modules/auth`
- `rg -n "Register\\(|/v1/auth|google" internal/transport/http/router internal/modules/auth -S`

3) Platform Google + token (jangan sebut file spesifik kalau belum pasti):
- `tree -L 4 internal/platform/google`
- `tree -L 4 internal/platform/token`
- `tree -L 4 internal/platform/state`
- `tree -L 4 internal/security`

---

## DECISIONS (jawab kalau belum jelas)
- Target client: web browser / API-to-API / keduanya?
- Token model: JWT access + refresh rotation (fixed) ✅
- Refresh cookie domain/samesite: `Lax` atau `Strict`? (default aman: Lax)
- Refresh TTL: 14d atau 30d?
- CSRF strategy untuk refresh/logout: double-submit + Origin allowlist? (default: iya)
- Minimal audit events yang wajib: login_start, login_success, login_fail, refresh_success, refresh_reuse_detected, logout, stepup_success, stepup_fail

---

## WORKFLOW (AI wajib ikuti)
1) Ringkas requirement final (bullet list).
2) Kritik risiko security + tradeoff (cookie/CSRF/origin/TTL/revoke).
3) Blueprint:
   - flow endpoint (start/callback/refresh/logout/stepup)
   - kontrak ports yang dipakai/ditambah
   - file list + ownership (modul mana pegang apa)
4) Eksekusi implementasi setelah blueprint disetujui.

---

## DELIVERABLES (saat eksekusi)
- Daftar file baru/berubah.
- Isi final tiap file (bukan potongan).
- Commands:
  - `gofmt -w .`
  - `go test ./... -count=1`
  - `go vet ./...`
  - `make audit`
- Curl sanity + expected output (minimal: health + error envelope + refresh behavior).
