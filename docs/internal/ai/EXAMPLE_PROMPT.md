# EXAMPLE_PROMPT (Auth Google)

### REPO CONTEXT
- Module: example.com/your-api
- Ikuti `docs/internal/ai/AI_RULES.md`
- Router: `internal/transport/http/router/*`
- Presenter: `internal/transport/http/presenter/*`
- Error response: JSON envelope via `presenter.HTTPErrorHandler`

### TASK
Buat fitur login/register user via Google OIDC, dengan:
- JWT access (30m)
- Refresh via HttpOnly cookie
- Refresh rotation + reuse detection
- Step-up (AAL2)

### REQUIRED SNAPSHOT
- `tree -L 6 internal/transport/http`
- `tree -L 6 internal/modules/auth`
- `cat internal/transport/http/router/v1/router.go`
- `cat internal/transport/http/presenter/error.go`
- `cat internal/platform/google/oidc_google.go`
- `rg -n "^package " internal/transport/http/router`

### DECISIONS
- Single-session: ya (login baru revoke lama)
- Audit events: login, logout, refresh, token_reuse_detected, stepup
