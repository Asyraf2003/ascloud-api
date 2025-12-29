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

---

## Database Transactions (Context-based)

Repo ini memakai pola transaksi berbasis `context.Context` supaya operasi multi-step bisa **atomic** tanpa “leak” SQL transaction ke layer usecase.

### Boundary rule (wajib)
- **Usecase hanya depend ke `ports.Transactor`** (tidak import `postgres`, tidak pegang `*sql.Tx`).
- Detail implementasi transaksi ada di layer platform/store (IO/vendor layer).

### Komponen
- `internal/platform/datastore/postgres/tx.go`
  - `WithTx(ctx, *sql.Tx)` menyisipkan transaksi ke context.
  - `GetExecutor(ctx, *sql.DB) SQLQueryer` mengembalikan executor:
    - `*sql.Tx` jika ada transaksi di context
    - `*sql.DB` jika tidak ada transaksi
  - `RunInTx(ctx, db, fn)` menjalankan `fn` dalam transaksi (commit/rollback).
- `internal/modules/auth/ports/transactor.go`
  - `ports.Transactor` = kontrak transaksi untuk usecase.
  - `ports.NoopTransactor` = implementasi untuk unit test (tanpa DB).
- Implementasi Postgres (platform)
  - `postgres.NewTransactor(db)` menghasilkan Transactor yang menjalankan `RunInTx`.

### Aturan Pakai
- **Usecase** yang melakukan beberapa operasi write harus dibungkus transaksi:
  - contoh: `GoogleCallback`: create account + link identity harus 1 transaksi.
- **Adapter Postgres (store/postgres)** wajib memakai `postgres.GetExecutor(ctx, db)` untuk semua query/exec, agar otomatis ikut transaksi jika ada.

### Anti-pattern (dilarang)
- Memulai transaksi dari repository/store (transaction boundary harus di usecase).
- Usecase mengimpor package platform (misal `internal/platform/datastore/postgres`) secara langsung.

### Kenapa begini?
- Menjamin **Atomicity (ACID)**: tidak ada data “setengah jadi”.
- Layering tetap rapi: usecase mengontrol “unit of work”, repo tetap fokus query.
- Mudah ganti implementasi DB di masa depan (selama kontrak executor terjaga).