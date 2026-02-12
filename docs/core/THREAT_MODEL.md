# Threat Model (AWS-First)

Dokumen ini adalah baseline threat model untuk repo ini.
Fokus: control-plane API + worker untuk hosting (statis dulu).

Jika ada perubahan auth model, trust boundary, atau hosting provisioning,
threat model ini wajib di-update melalui ADR.

---

## Assumptions
- Client utama: web/dashboard (repo terpisah) yang consume API.
- Auth: JWT access + refresh rotation (refresh idealnya lewat HttpOnly cookie).
- API dan worker deploy terpisah (Lambda).
- Vendor/IO hanya lewat `internal/platform/*` dan diakses via `ports`.
- Setiap resource wajib scoped ke tenant/account (anti IDOR).
- Log tidak boleh memuat secrets (Authorization/Cookie/token).
- Baseline infra: AWS (CloudFront, S3, SQS, DynamoDB, Lambda).

---

## Assets (yang harus dilindungi)
- Account identity (user, tenant)
- Session/refresh tokens (refresh hash + session state + anti-reuse signals)
- Domain ownership proof (DNS/verification state)
- Hosting artifacts (site releases, immutable outputs)
- Job state (idempotency, attempts, failure reasons)
- Usage/quota counters (storage/pageview)
- Audit logs (integrity + append-only semantics)
- Config + secrets runtime (ENV, IAM credentials/roles)

---

## Trust Boundaries
- Client (web/dashboard) ↔ API (HTTPS; cookie refresh boundary)
- API ↔ DynamoDB (metadata)
- API ↔ S3 (pre-signed URL issuance)
- API ↔ SQS (enqueue job)
- Worker ↔ SQS (consume job)
- Worker ↔ S3 (download uploads + publish releases)
- Edge plane:
  - Visitor ↔ CloudFront
  - CloudFront Function ↔ KVS (routing decision)
  - CloudFront ↔ S3 (origin)
- External IdP (Google OIDC) ↔ API

### Legacy boundary (historical)
- API ↔ Postgres (legacy_postgres tag only). Ini bukan baseline.

---

## Threats & Mitigations (baseline)

Kolom Owner itu “pemilik area kode” yang wajib enforce.

| Threat | Example | Mitigation (baseline) | Owner |
|---|---|---|---|
| Token theft | Token bocor dari log / XSS / extension | Redact log ketat, HttpOnly cookie untuk refresh, TLS, short-lived access JWT | `presenter`, auth module |
| Refresh reuse / replay | Refresh lama dipakai ulang | Rotation + reuse detection, revoke session, audit event | auth usecase/store |
| CSRF (cookie refresh) | Cross-site POST refresh | Origin allowlist + CSRF double-submit, SameSite cookie sesuai mode | middleware trust + auth usecase |
| Broken access control (IDOR) | Ambil resource tenant lain | Scoped queries by owner/tenant, explicit authorization checks per usecase | semua usecase modul |
| Brute force / abuse | Spam login/callback/upload/job | Rate limiting plan + trust scoring + audit suspicious attempts | trust module + middleware |
| ZIP slip / path traversal | ZIP berisi `../` | Validate entry path, reject traversal, normalize | worker (hosting pipeline) |
| ZIP bomb / resource exhaustion | Millions files / huge decompression | Limit total extracted bytes, file count, depth, timeout | worker |
| Symlink abuse | ZIP membuat symlink untuk escape | Reject symlink/hardlink entries | worker |
| Artifact exfiltration | Object public kebuka | Bucket private by default, signed URLs only untuk upload, least privilege IAM | objectstore adapter + infra |
| Cache poisoning / wrong content-type | HTML diserve sebagai js / sniff | Set Content-Type, `nosniff`, cache policy (HTML short TTL, assets long TTL) | worker + edge policy |
| Host routing abuse | Host header mapped ke site lain | KVS mapping validated, host allowlist/ownership enforced, no wildcard careless | domains/hosting usecase + edge sync |
| Data leakage via errors | AWS SDK error mentah ke user | AppError mapping + sanitize in presenter; worker simpan error_code safe | presenter + apperr |
| Audit log tampering | Mengubah event historis | Append-only semantics (DynamoDB patterns / stream), minimal update permissions | audit module + datastore |
| Misconfig (CORS/origin) | Origin terlalu permisif | Explicit allowlist, config validation, environment-specific defaults | config + middleware |

---

## Open Risks / TODO (tracked)
- Observability: metrics/tracing + security signals (auth anomalies, job failure spikes)
- Incident response: runbook + alerting + escalation path
- Rate limit plan: definisikan policy final (per IP, per account, per endpoint)
- Dynamic hosting sandbox (future): isolation boundary, policy, quotas
