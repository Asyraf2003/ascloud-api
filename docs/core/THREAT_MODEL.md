# Threat Model

Dokumen ini adalah baseline threat model untuk repo ini. Fokus: control-plane API + worker untuk hosting (statis dulu).
Kalau ada perubahan auth model, trust boundary, atau hosting provisioning, threat model ini wajib di-update.

---

## Assumptions (yang dianggap benar)
- Client utama: web/dashboard (repo terpisah) yang consume API.
- Auth: JWT access + refresh rotation (refresh idealnya lewat HttpOnly cookie).
- API dan worker deploy terpisah.
- Vendor/IO hanya lewat `internal/platform/*` dan diakses via `ports`.
- Setiap resource wajib scoped ke tenant/account (anti IDOR).
- Log tidak boleh memuat secrets (Authorization/Cookie/token).

---

## Assets (yang harus dilindungi)
- Account identity (user, tenant)
- Session/refresh tokens (termasuk refresh hash dan session state)
- Domain ownership proof (DNS/verification state)
- Hosting artifacts (files/build outputs)
- Billing data (jika ada)
- Audit logs (integrity + append-only semantics)
- Config + secrets runtime (ENV, credentials platform)

---

## Trust Boundaries
- Client (web/dashboard) ↔ API
- API ↔ DB (Postgres)
- Worker ↔ Queue (SQS/local)
- Worker ↔ Objectstore (S3/R2)
- Worker ↔ Edge provider (Cloudflare/CloudFront)
- External IdP (Google OIDC) ↔ API

---

## Threats & Mitigations (baseline)

Kolom Owner itu bukan “orang”. Itu pemilik area kode yang wajib enforce.

| Threat | Example | Mitigation (baseline) | Owner |
|---|---|---|---|
| Token theft | Token bocor dari log / XSS / extension | Redact log ketat, HttpOnly cookie untuk refresh, TLS, short-lived access JWT | `internal/transport/http/presenter`, `auth` module |
| Session replay / refresh reuse | Refresh lama dipakai ulang | Refresh rotation + reuse detection, revoke session, audit event | `internal/modules/auth/usecase`, auth store |
| CSRF (cookie-based refresh) | Cross-site POST refresh | Origin allowlist + CSRF double-submit, SameSite cookie (sesuai kebutuhan) | `internal/transport/http/middleware`, `auth` usecase |
| Brute force / credential stuffing | Spam endpoint login/callback | Rate limiting, trust scoring, backoff, audit suspicious attempts | `trust` module + middleware, (future) rate-limit layer |
| Broken access control (IDOR) | Ambil resource tenant lain | Scoped queries by tenant/account, explicit authorization checks per usecase | Semua `internal/modules/*/usecase` (owner per modul) |
| SSRF dari integration | Edge/objectstore callback memicu request internal | Allowlist host/endpoint, block private ranges, validate URLs, no raw fetch | `internal/platform/*` adapters + `hosting`/`domains` usecase |
| Supply chain risk | Dependency malicious / CI compromise | `go.sum` committed, pin versions, CI checks, secret scanning, least privilege CI | CI workflows + repo policy |
| Audit log tampering | Mengubah event historis | Append-only table/stream, immutability policy, checksum/constraints, minimal update permissions | `audit` module + datastore schema |
| Data exfiltration (artifacts) | Public object leakage | Private buckets by default, scoped access, signed URLs, least privilege IAM | `internal/platform/objectstore/*` + `hosting` module |
| Misconfig (CORS/origin) | Origin terlalu permisif | Explicit allowlist, environment-specific config validation | `internal/config` + middleware |
| Sensitive data in error responses | SQL/vendor error bocor ke user | Error envelope sanitization via presenter, AppError mapping | `internal/transport/http/presenter`, `internal/shared/apperr` |

---

## Open Risks / TODO (tracked)
- Observability: metrics/tracing + security signals (auth anomalies, rate-limit hits)
- Incident response: runbook + alerting + escalation path
- Dynamic hosting sandbox (future): isolation boundary (container/VM), policy, quotas
- Rate limit plan: definisikan policy final (per IP, per account, per endpoint)
