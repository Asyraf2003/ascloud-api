# Module: auth

Scope:
- Login/Register via Google OIDC
- Session + refresh rotation + reuse detection
- Issue JWT access + session-check per request (sid must be active)
- Step-up (AAL2) hooks
- Audit sink hooks + trust evaluation hooks

Boundaries (hexagonal):
- domain/  : entities + invariants + semantic errors
- ports/   : interfaces (OIDCProvider, SessionStore, IdentityRepo, TokenIssuer, StateStore, TrustEvaluator, AuditSink, AccountService)
- usecase/ : orchestration flows (start/callback/refresh/logout)
- transport/http/ : request mapping + light validation + call usecase (responses via presenter only)

HTTP endpoints (v1):
- GET  /v1/auth/google/start
- GET  /v1/auth/google/callback
- POST /v1/auth/refresh
- POST /v1/auth/logout
- (future) /v1/auth/google/stepup/start + callback

Security notes:
- Refresh token in HttpOnly cookie, rotated on refresh
- Reuse detection: reuse -> revoke session + audit event
- Sensitive data never returned; errors via presenter.HTTPErrorHandler only

Testing:
- Unit tests: usecase/domain with fakes via ports
- Component tests: httptest router/middleware/handlers with build tag `component`

TODO:
- Hardening step-up AAL2 flow endpoints
- Observability: metrics/tracing + audit drill
