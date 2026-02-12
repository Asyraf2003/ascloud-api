# Structure & Contracts (AWS-First)

Dokumen ini adalah pegangan struktur repo dan kontrak antar layer.
Tujuan: gampang dibaca, gampang diaudit, minim efek berantai saat perubahan.

> Hard rules paling detail ada di `docs/internal/ai/AI_RULES.md`.

---

## Scope Produk
- Sistem inti: **API + Worker**.
- Web/dashboard berada di repo lain (client) dan hanya consume API.
- Fokus launch awal: **hosting statis**.
- Baseline vendor aktif: **AWS** (CloudFront, S3, SQS, DynamoDB, Lambda).

---

## Layout Repo

### `cmd/`
Entrypoint binary.
- `cmd/api`: HTTP API (control-plane)
- `cmd/worker`: job async (publish/sync/reconcile)
- `cmd/migrate`: **legacy Postgres tooling** (jalan dengan `-tags=legacy_postgres`)

### `internal/modules/<module>/`
Modul bisnis per bounded-context.

Struktur standar:
- `domain/` : entity + invariants
- `ports/`  : interface dependency
- `usecase/`: orchestration flow
- `transport/http/`: HTTP adapter khusus modul
- `store/...` (opsional): adapter storage spesifik modul (contoh: legacy postgres)

### `internal/transport/http/`
HTTP “core” lintas modul:
- `router/`, `middleware/`, `presenter/`

### `internal/platform/`
Adapter vendor/IO.
Baseline AWS-first:
- datastore: `internal/platform/datastore/dynamodb` (aktif)
- objectstore: `internal/platform/objectstore/s3` (aktif)
- queue: `internal/platform/queue/sqs` (aktif)
- edge: `internal/platform/edge/cloudfront` (aktif)

Adapter lain di repo (cloudflare/r2/postgres) adalah **inaktif/legacy** kecuali ada ADR baru.

### Lainnya
- `internal/config/`: load & validate konfigurasi
- `internal/app/bootstrap/`: wiring/DI
- `deploy/`: dev tooling (docker, dsb)
- `scripts/`: tooling audit

### Legacy folder
- `migrations/`: **legacy SQL migrations (Postgres)**, bukan baseline AWS-first.
- `internal/platform/datastore/postgres`: legacy (tagged)
- `internal/modules/*/store/postgres`: legacy (tagged)

---

## Contracts Antar Layer (ringkas)
(isi bagian kontrak kamu yang sekarang sudah bagus, bisa tetap)
- Domain: no platform/http/module lain
- Ports: interface only, minimal deps
- Usecase: domain + ports + shared only
- Transport/http: adapter, tidak boleh import platform
- Platform: vendor adapter, tidak boleh import http

Enforcement: `scripts/audit_boundaries.sh`

---

## Repo Audits (Quality Gates)
- `make audit`
- `bash scripts/audit_boundaries.sh`
- `bash scripts/docs_audit.sh`
- `bash scripts/testtags_audit.sh`
- `bash scripts/content_audit.sh`

---

## Legacy Policy
Legacy Postgres tidak boleh mempengaruhi default build.
Semua legacy harus:
- diisolasi dengan build tags (`legacy_postgres`)
- diberi prefix `legacy_*`
- tidak menjadi rujukan baseline docs (kecuali bagian “Legacy”)
