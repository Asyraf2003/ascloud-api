# EXAMPLE_PROMPT

Contoh prompt yang sudah terisi (realistic), mengikuti format `AI_PROMPT.md`.

---

## REPO CONTEXT
- Module: example.com/your-api
- Ikuti `docs/internal/ai/AI_RULES.md` (hard rules + boundaries + DoD).
- Router: `internal/transport/http/router/*` (router induk + v1 modular).
- Presenter: `internal/transport/http/presenter/*`.
- Error response: JSON envelope via `presenter.HTTPErrorHandler` (no secrets).
- Debug routes wajib gated: `DEBUG_ROUTES=1`.
- 1 folder = 1 package. File <= 100 baris.

## TASK
Buat blueprint + implementasi **login/register via Google OIDC** (tanpa metode lain), dengan:
- Access token: JWT (exp 30 menit)
- Refresh token: HttpOnly cookie
- Refresh rotation + anti-reuse detection
- Single-session per user (login baru revoke session lama)
- Step-up auth (AAL2) untuk operasi sensitif
- Audit events minimal untuk semua event auth

## REQUIRED SNAPSHOT (WAJIB, paste output)
1) Struktur:
- `tree -L 6 internal/transport/http`
- `tree -L 6 internal/modules/auth`

2) Kontrak inti:
- `cat internal/transport/http/router/router.go`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `rg -n "^package " internal/transport/http/router`

3) Extra snapshot (karena auth + google + token):
- `tree -L 6 internal/platform/google`
- `tree -L 6 internal/platform/token`
- `cat internal/platform/google/oidc_google.go`
- `cat internal/platform/google/idtoken_verifier.go`
- `cat internal/platform/token/jwt/issuer.go`
- `cat internal/modules/auth/usecase/service.go` (kalau ada)
- `rg -n "refresh|rotation|csrf|origin|aal|stepup" internal/modules/auth -S`

## DECISIONS (confirm sebelum mulai)
- Target client: web browser (cookie-based) → butuh CORS/Origin strict + CSRF double-submit.
- Token delivery: refresh via HttpOnly cookie (bukan response body).
- Expiry: access 30m, refresh 30d (atau 14d, pilih satu konsisten).
- Protected routes: wajib session-check (sid masih aktif), karena JWT stateless.
- Audit events minimal: login_start, login_success, login_failed, refresh_rotated, refresh_reused, logout, stepup_success, stepup_failed.

## WORKFLOW (AI wajib ikuti)
1) Ringkas requirement final (bullet list) dari TASK + keputusan di atas.
2) Kritik bagian berisiko (cookie + CSRF, session-check cost, replay risk) dan tawarkan mitigasi.
3) Buat blueprint:
   - endpoint list + flow ringkas
   - kontrak ports (OIDCProvider, SessionStore, TokenIssuer, AuditSink, TrustEvaluator)
   - file list + ownership (domain/ports/usecase/transport/http/platform)
4) Setelah blueprint disetujui, baru implementasi.

## DELIVERABLES (saat eksekusi)
- Daftar file baru/berubah.
- Isi final tiap file (bukan potongan).
- Command:
  - `gofmt -w .`
  - `go test ./... -count=1`
  - `go vet ./...`
  - `make audit`
- Curl sanity + expected response:
  - `GET /health` → 200
  - `GET /v1/health` → 200
  - `GET /ga-ada` → 404 JSON envelope
  - Auth flows: start redirect, callback success, refresh rotate, reuse detection triggers revoke
