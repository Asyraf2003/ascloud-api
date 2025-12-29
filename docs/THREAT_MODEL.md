# Threat Model

## Assets (yang harus dilindungi)
- Account identity (user, tenant)
- Session/refresh tokens
- Domain ownership proof
- Hosting artifacts (files/build outputs)
- Billing data (jika ada)
- Audit logs (integrity)

## Trust Boundaries
- Client (web/dashboard) ↔ API
- API ↔ DB (Postgres)
- Worker ↔ Queue (SQS/local)
- Worker ↔ Objectstore (S3/R2)
- Worker ↔ Edge provider (Cloudflare/CloudFront)
- External IdP (Google OIDC) ↔ API

## Threats (contoh awal)
- Token theft / session replay
- CSRF (jika cookie dipakai)
- Brute force / credential stuffing
- Broken access control (IDOR)
- SSRF dari provisioning/edge integration
- Supply chain risk (deps, CI)
- Data tampering pada audit logs

## Mitigations (baseline)
- Refresh rotation + revoke + device metadata
- Rate limiting + trust scoring + origin checks
- Strict authorization per resource (tenant/account scoping)
- Logging terstruktur + request_id + redact secrets
- Least privilege IAM untuk objectstore/queue/edge
- Migration & schema constraints untuk integrity

## Open Risks / TODO
- Observability (metrics/tracing)
- Incident response + runbook
- Dynamic hosting sandbox (future)
