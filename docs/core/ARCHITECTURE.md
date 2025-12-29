# Architecture

Dokumen ini menjelaskan peta sistem (high-level) untuk repo **your-api**.
Tujuan: orang bisa memahami “ini sistem apa, komponen apa saja, jalur request gimana, dan batas tanggung jawabnya” tanpa baca seluruh kode.

**Referensi**
- Contracts & boundaries: `docs/core/STRUCTURE.md`
- Threat model: `docs/core/THREAT_MODEL.md`
- ADR: `docs/adr/*`

---

## Goal
Launch **hosting statis** lebih dulu dengan fondasi yang siap berkembang menjadi **hosting dinamis** lewat pola API + Worker.

---

## Scope (Repo ini)
Repo ini berisi:
- `cmd/api`: HTTP API (control-plane)
- `cmd/worker`: job async (publish/sync/reconcile)
- `cmd/migrate`: migrasi schema

Repo ini **tidak** berisi web/dashboard. Web ada di repo lain sebagai client yang hanya consume API.

---

## Non-goals (Saat ini)
- Tidak mengejar microservices banyak sejak awal.
- Tidak membangun runtime dinamis dulu (container/serverless/managed runtime baru fase berikutnya).
- Tidak menjadikan repositori ini tempat “catatan random” (itu tempatnya `docs/notes/`).

---

## High-level Design
Model: **modular monolith** (1 repo, boundary ketat), dengan **2 binary deployable**:
- API (sync request/response)
- Worker (async jobs)

Setiap modul mengikuti pola:
`domain` → `ports` → `usecase` → adapter (`transport/http`, `store/*`, `platform/*`)

Prinsip utama:
- Domain/usecase tidak boleh bergantung pada platform/IO secara langsung.
- IO/vendor ada di `internal/platform/*` dan masuk melalui interface `ports`.

---

## Runtime Components

### Control-plane API (`cmd/api`)
Tugas:
- Auth & session orchestration
- Endpoint manajemen account/domain/hosting
- Validasi request, enforcement middleware, response envelope

Peta:
- Router: `internal/transport/http/router`
- Middleware: `internal/transport/http/middleware`
- Presenter: `internal/transport/http/presenter`
- Modul handlers: `internal/modules/*/transport/http`

### Worker (`cmd/worker`)
Tugas:
- Menjalankan job async:
  - publish static artifact ke objectstore
  - sync edge (Cloudflare/CloudFront)
  - reconcile/renewal (future)
  - provisioning runtime dinamis (future)

Worker mengakses vendor lewat:
- `internal/platform/queue/*`
- `internal/platform/objectstore/*`
- `internal/platform/edge/*`
- datastore via `internal/platform/datastore/*`

---

## System Topology (Mental Model)

Client (web/dashboard)
  ↕ HTTPS
API (Echo) ──↔ Postgres (sessions, identities, accounts, audit, ...)
   │
   ├── publish job → Queue (local/SQS)
   │                 ↕
   │               Worker ──→ Objectstore (S3/R2)
   │                 │
   │                 └────→ Edge (Cloudflare/CloudFront)

External IdP:
- Google OIDC ↔ API

Catatan: detail security boundary dan mitigasi lihat `docs/core/THREAT_MODEL.md`.

---

## Module Map (Current)
- `auth`: login/register (Google OIDC), session, refresh rotation, step-up (AAL), trust hooks, audit sink
- `account`: account/user/tenant (placeholder untuk ownership)
- `domains`: domain management + verification (future)
- `hosting`: hosting project/site (static dulu)
- `trust`: trust scoring + policy enforcement hooks
- `audit`: audit events (append-only)
- `system`: health/debug endpoints

---

## Key Flows

### 1) Auth (Google OIDC)
Ringkas:
- `/v1/auth/google/start` → buat state/nonce/PKCE (TTL pendek) → redirect ke Google
- `/v1/auth/google/callback` → verify id_token → resolve identity (provider+sub) → create/link account → create session → issue JWT access + set refresh cookie
- `/v1/auth/refresh` → CSRF double-submit + Origin allowlist → refresh rotation + anti-reuse → issue access JWT baru
- `/v1/auth/logout` → revoke session + clear cookies
- Step-up flow untuk AAL2 (future-hardening)

Sumber keputusan: `docs/adr/0001-auth-google-oidc.md`

### 2) Hosting v1: Static (Launch)
Target flow:
1) Client create project/site → API (`hosting` usecase)
2) API enqueue job publish → Worker
3) Worker publish artifact ke objectstore (S3/R2)
4) Worker configure edge (Cloudflare/CloudFront)
5) Domain attach/verify → `domains` usecase + worker sync (future)

Catatan: implementasi detail static hosting akan ditulis via ADR saat mulai dikerjakan.

### 3) Hosting v2: Dynamic (Future)
Konsep (tanpa implementasi dulu):
- Tambah port seperti `RuntimeProvisioner` (create/update/delete runtime)
- Worker jadi eksekutor provisioning
- Hosting module tetap abstrak (tidak mengunci ke AWS detail)

Saat keputusan konkret diambil (container vs serverless vs managed runtime), wajib ADR baru.

---

## Data Ownership (Rule)
- Setiap modul “memiliki” data/invariants-nya.
- Dilarang join/random query lintas modul di domain/usecase.
- Kebutuhan lintas modul wajib lewat port explicit (service/repo interface), bukan import paket modul lain.

Enforcement boundary: `scripts/audit_boundaries.sh`

---

## Quality Gates (Target 9/10)
Repo dianggap sehat kalau:
- `make audit` lulus (unit + component + vet + boundary + style checks)
- Tidak ada doc yang “nyasar” (ADR harus ADR, notes harus notes)
- Keputusan besar punya ADR
- Threat model minimal di-update saat:
  - auth model berubah
  - trust boundary berubah
  - hosting dinamis mulai dikerjakan
