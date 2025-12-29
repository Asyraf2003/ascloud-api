# Architecture

Dokumen ini menjelaskan peta sistem (high-level) untuk repo **your-api**.
Tujuan: orang paham “ini sistem apa, komponen apa saja, alur request seperti apa, dan batas tanggung jawabnya” tanpa baca semua kode.

## Referensi
- Structure & contracts: `docs/core/STRUCTURE.md`
- Threat model: `docs/core/THREAT_MODEL.md`
- ADR: `docs/adr/*`

---

## Goal
Launch **hosting statis** lebih dulu, dengan fondasi yang siap berkembang ke **hosting dinamis** melalui pola **API + Worker**.

---

## Scope (Repo ini)
Repo ini berisi:
- `cmd/api`: HTTP API (control-plane)
- `cmd/worker`: job async (publish/sync/reconcile)
- `cmd/migrate`: migrasi schema

Repo ini **tidak** berisi web/dashboard. Web ada di repo lain sebagai client yang hanya consume API.

---

## Non-goals (Saat ini)
- Tidak mengejar banyak microservices sejak awal.
- Tidak membangun runtime dinamis (container/serverless/managed runtime) dulu.
- Tidak menaruh catatan random di core docs (catatan masuk `docs/notes/`).

---

## Glossary (biar nggak debat istilah)
- **Control-plane**: API yang mengelola state dan orchestration (auth, domain, hosting config, enqueue job).
- **Worker**: eksekutor async untuk hal IO-heavy (publish artifact, edge sync, reconcile).
- **Artifact**: output hosting statis (file/site build) yang dipublish ke objectstore lalu diserve via edge.
- **Job**: unit kerja idempotent yang diproses worker (dipush via queue).

---

## High-level Design
Model: **modular monolith** (1 repo, boundary ketat), dengan **2 binary deployable**:
- API (sync request/response)
- Worker (async jobs)

Pola modul:
`domain` → `ports` → `usecase` → adapter (`transport/http`, `store/*`, `platform/*`)

Prinsip:
- Domain/usecase **tidak** boleh bergantung pada IO/vendor.
- IO/vendor ada di `internal/platform/*` dan masuk via interface `ports`.

---

## Runtime Components

### Control-plane API (`cmd/api`)
Tugas:
- Auth & session orchestration
- Endpoint manajemen account/domain/hosting
- Validasi request ringan + middleware enforcement + response envelope

Peta layer:
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

Akses vendor via:
- Queue: `internal/platform/queue/*`
- Objectstore: `internal/platform/objectstore/*`
- Edge: `internal/platform/edge/*`
- Datastore: `internal/platform/datastore/*`

---

## System Topology (Mental Model)

```text
Client (web/dashboard)
   ↕ HTTPS
API (Echo; control-plane) ──↔ Postgres (sessions, identities, accounts, audit, ...)
   │
   ├── enqueue job → Queue (local/SQS)
   │                  ↕
   │                Worker ──→ Objectstore (S3/R2) ──→ Edge (Cloudflare/CloudFront)
   │
External IdP:
- Google OIDC ↔ API
