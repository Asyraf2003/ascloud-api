# Architecture (AWS-First) — Static Hosting via ZIP

Dokumen ini adalah peta sistem (high-level) untuk repo **your-api**.
Tujuan: orang paham “ini sistem apa, komponen apa saja, alur utama seperti apa, dan batas tanggung jawabnya”
tanpa baca semua kode.

Dokumen ini menggabungkan:
- Peta sistem (architecture map)
- Baseline blueprint (AWS-first, serverless, immutable releases)

Jika ada konflik dengan diskusi lain atau catatan lama: ikuti dokumen ini + ADR terbaru.

---

## Referensi
- Structure & contracts: `docs/core/STRUCTURE.md`
- Threat model: `docs/core/THREAT_MODEL.md`
- Error handling: `docs/core/ERROR_HANDLING.md`
- ADR: `docs/adr/*`

---

## Goal
Launch **hosting statis** lebih dulu, dengan fondasi siap berkembang ke **hosting dinamis**
melalui pola **API + Worker** dan pendekatan event-driven.

---

## Keputusan Baseline (MVP Pegangan, Non-Negotiable)
- **Provider aktif:** AWS only (yang lain di repo bersifat persiapan/inaktif sampai ada keputusan eksplisit baru)
- **CDN & Edge:** CloudFront + CloudFront Function + KeyValueStore (KVS)
- **Storage:** S3
- **Compute:** Lambda (API + Worker)
- **Queue:** SQS (Standard)
- **Metadata DB:** DynamoDB (on-demand)
- **Region:** ap-southeast-3 (Jakarta)
- **Release model:** Immutable release + pointer rollback
- **Storage quota:** 20MB extracted size per site, **HARD FAIL** deploy jika lewat
- **Pageview quota:** HTML-only dari CloudFront logs (batch), **SOFT WARNING + GRACE** sebelum suspend
- **Auth browser:** Dashboard & API 1 eTLD+1 (mis `app.domain.com` dan `api.domain.com`) dengan refresh token rotation via HttpOnly cookie + mitigasi CSRF
- **EC2:** dilarang untuk MVP

---

## Scope (Repo ini)
Repo ini berisi:
- `cmd/api`: HTTP API (control-plane)
- `cmd/worker`: job async (publish/sync/reconcile)
- `cmd/migrate`: **legacy tooling** (lihat bagian Legacy)

Repo ini **tidak** berisi web/dashboard. Web ada di repo lain sebagai client yang consume API.

---

## Glossary (biar nggak debat istilah)
- **Control-plane**: API yang mengelola state dan orchestration (auth, domain, hosting config, enqueue job).
- **Worker**: eksekutor async untuk hal IO-heavy (extract ZIP, publish ke S3, sync edge).
- **Artifact**: output hosting statis yang dipublish ke S3 release folder.
- **Release**: versi immutable dari site. Setiap deploy membuat release baru.
- **Pointer rollback**: rollback hanya mengubah `current_release_id`, tidak memindahkan file.
- **Job**: unit kerja idempotent diproses worker (dipush via queue).

---

## High-level Design
Model: **modular monolith** (1 repo, boundary ketat), dengan 2 binary deployable:
- API (sync request/response)
- Worker (async jobs)

Pola modul:
`domain` → `ports` → `usecase` → adapter (`transport/http`, `store/*`, `platform/*`)

Prinsip:
- Domain/usecase **tidak** boleh bergantung pada IO/vendor.
- IO/vendor ada di `internal/platform/*` dan masuk via interface `ports`.
- Kontrak publik (router/presenter/error handler) dijaga ketat (lihat ADR + AI_RULES).

---

## Runtime Components

### 1) Control-plane API (`cmd/api`)
Tugas:
- Google OIDC login
- Session: refresh token rotation (HttpOnly cookie) + anti-reuse
- Generate pre-signed S3 upload URL
- Create site/release/job metadata
- Enqueue SQS job
- List release, rollback (update pointer `current_release_id`)
- Quota/status site (ACTIVE / SUSPENDED)

Peta layer:
- Router: `internal/transport/http/router`
- Middleware: `internal/transport/http/middleware`
- Presenter: `internal/transport/http/presenter`
- Modul handlers: `internal/modules/*/transport/http`

### 2) Worker (`cmd/worker`)
Tugas:
- Proses job async dari SQS
- Download ZIP dari S3 (uploads)
- **ZIP safety** (non-negotiable): anti zip-slip, anti zip-bomb (limit size/file/depth), reject symlink, timeout
- Extract ke folder release immutable:
  `sites/{site_id}/releases/{release_id}/...`
- Hitung total size hasil extract
  - jika > 20MB → FAIL deploy
- Set content-type & cache headers
- Update status job + release di DynamoDB
- Hapus ZIP upload (best effort)

Akses vendor via:
- Queue: `internal/platform/queue/sqs` (aktif), `internal/platform/queue/local` (dev)
- Objectstore: `internal/platform/objectstore/s3` (aktif)
- Edge: `internal/platform/edge/cloudfront` (aktif)
- Datastore: `internal/platform/datastore/dynamodb` (aktif)

---

## System Topology (Mental Model)

Client (web/dashboard)
   ↕ HTTPS (1 eTLD+1; cookie refresh)
API (Lambda; Echo; control-plane) ──→ DynamoDB (metadata)
   │
   ├── create upload url → S3 (uploads/{job_id}.zip)
   │
   └── enqueue job → SQS
                    ↕
                  Worker (Lambda; SQS trigger)
                    ├── read ZIP from S3 uploads
                    ├── validate + extract → S3 releases (immutable)
                    ├── update DynamoDB (job/release/site pointer)
                    └── sync edge metadata (KVS) (host→site_id / site→current_release / suspended)

Visitor (public traffic)
   ↕ HTTPS
CloudFront → Function (rewrite via KVS) → S3 release folder

External IdP:
- Google OIDC ↔ API

---

## Data Model (DynamoDB-Friendly)
Minimum baseline (bisa evolve via ADR):
- **users**: `user_id` (PK), `google_sub` (unique), `email`, `created_at`
- **sites**: `site_id` (PK), `owner_user_id`, `slug` (unique), `current_release_id`, `status` (ACTIVE/SUSPENDED), `created_at`
- **releases**: `release_id` (PK), `site_id`, `status` (PENDING/SUCCESS/FAILED), `size_bytes`, `error_code`, `error_reason`, `created_at`
- **jobs**: `job_id` (PK), `site_id`, `release_id`, `zip_key`, `status` (QUEUED/RUNNING/SUCCESS/FAILED), `attempts`, `created_at`
- **usage_monthly**: `site_id#yyyymm` (PK), `storage_bytes`, `pageviews_html`, `updated_at`

---

## Main Flows

### Deploy (static)
1. API: create release + job (status PENDING/QUEUED), return pre-signed upload URL.
2. Client: upload ZIP → S3 `uploads/{job_id}.zip`
3. API: enqueue SQS job (idempotency key recommended).
4. Worker: validate ZIP + extract → S3 `sites/{site_id}/releases/{release_id}/...`
5. Worker: enforce 20MB extracted size (HARD FAIL if > limit).
6. Worker: update release/job status.
7. API: set `sites.current_release_id = release_id` (atau worker bisa set, sesuai ADR flow).
8. Edge: KVS mapping memastikan host → current release.

### Rollback
1. API update `sites.current_release_id` ke release sebelumnya (pointer rollback).
2. Edge mapping ikut update (KVS).
3. Tidak ada file movement dan tidak rebuild.

---

## Edge Routing (CloudFront Function + KVS)
KVS menyimpan mapping minimal:
- `host` → `site_id`
- `site_id` → `current_release_id`
- `site_id` → `suspended` flag

CloudFront Function melakukan rewrite internal:
`/{path}` → `/sites/{site_id}/releases/{current_release_id}/{path}`

Catatan:
- Sanitasi path (no traversal)
- Default ke `/index.html` bila path folder (future detail via ADR)

---

## Quota & Enforcement
- **Storage quota (Hard):** 20MB extracted size per site. Over limit → deploy FAIL.
- **Pageview (Soft):** hitung HTML request dari CloudFront logs (batch).
  - Warning → grace period → suspend
- Suspend = update `sites.status` + set flag di KVS (stop serving atau serve “suspended page”, sesuai kebijakan).

---

## Security Baseline (ringkas)
- Refresh token: hashed at rest, rotation, reuse detection, revoke
- CSRF mitigation untuk endpoint cookie-based
- Log redaction (Authorization/Cookie/token)
- ZIP safety ketat (zip-slip/bomb/symlink/depth/timeout)
- Least privilege IAM untuk Lambda/Worker
- Jangan bocorin vendor error mentah ke client (lihat ERROR_HANDLING)

---

## Observability & Audit
- Logs terstruktur minimal: `request_id` (API), `job_id`, `site_id`, `release_id` (Worker)
- Metrics: job success rate, queue depth, deploy duration, Lambda error rate
- Audit event minimal: deploy/rollback/auth security events

---

## Legacy (dibekukan, tidak jadi baseline)
Repo ini masih menyimpan jejak Postgres lama **untuk histori dan referensi**, tapi:
- Default build **tidak** boleh bergantung Postgres / `database/sql`
- Legacy hanya berjalan dengan build tag:
  - `-tags=legacy_postgres`

Semua hal legacy harus:
- diberi prefix/penamaan `legacy_*`
- tidak mempengaruhi default path (unit/component tests default tetap aman)
- tidak mengubah keputusan baseline tanpa ADR baru
