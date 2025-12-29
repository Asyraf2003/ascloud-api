Contoh Prompt Terisi (contoh: login Google)
KONTEKS REPO
- Module: example.com/your-api
- Aturan repo: ikuti docs/AI_RULES.md.
- Router: internal/transport/http/router/* (router induk + v1 modular).
- Presenter: internal/transport/http/presenter/*.
- Error response JSON envelope via presenter.HTTPErrorHandler. JSONB hanya storage internal.

TASK
- Buat blueprint login/register user via Google ID Token (tanpa metode lain).
- Struktur harus siap untuk tambah provider lain nanti (Apple, Email, dsb) tanpa bongkar usecase.

SEBELUM MULAI (WAJIB)
1) Minta snapshot repo:
   - tree -L 6 internal/transport/http
   - tree -L 6 internal/modules/auth
   - cat internal/transport/http/router/router.go
   - cat internal/transport/http/router/v1/router.go
   - cat internal/transport/http/presenter/error.go
   - rg -n "^package " internal/transport/http/router
   - cat internal/platform/google/idtoken_verifier.go
   - cat internal/security/token/token.go

2) Tanyakan keputusan:
   - Target client: web atau mobile?
   - Refresh token mau di cookie HttpOnly atau return di body?
   - Access token expiry: 10m atau 15m?
   - Refresh token expiry: 14d atau 30d?
   - Audit events minimal apa saja?

WORKFLOW
- Setelah jawaban ada:
  - ringkas requirement, kritik, blueprint, minta saya audit, baru eksekusi.

DELIVERABLES
- file list, final file content, commands test, curl sanity tests, minim efek berantai, file <=