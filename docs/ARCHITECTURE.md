# Architecture

## Goal
Membangun layanan API untuk hosting statis terlebih dahulu (launch awal), dengan fondasi yang siap diperluas ke hosting dinamis lewat API + worker.

## Scope
- Repo ini berisi:
  - `cmd/api`: HTTP API (control-plane)
  - `cmd/worker`: job async (provisioning/sync/billing ticks)
  - `cmd/migrate`: migrasi schema
- Web/dashboard berada di repo lain (client) dan hanya consume API.

## High-level Design
- Modular monolith (satu repo, boundary ketat), deployable terpisah untuk API dan worker.
- Per modul: `domain` → `ports` → `usecase` → adapter (`transport/store`).
- Prinsip: domain/usecase tidak boleh bergantung pada transport/platform secara langsung.

## Module Map (Current)
- `auth`: login, session, refresh rotation, trust hooks, audit sink
- `account`: akun user/tenant
- `domains`: domain management
- `hosting`: hosting projects/sites (statis dulu)
- `trust`: evaluasi trust & policy enforcement
- `audit`: audit events
- `system`: health/debug endpoints

## Runtime Components
### API (`cmd/api`)
- Router: `internal/transport/http/router`
- Middleware: request_id, access_log, auth_context, jwt_auth, trust/*

### Worker (`cmd/worker`)
- Mengkonsumsi queue / menjalankan job:
  - publish static
  - sync edge
  - renewal / reconciliation
  - (future) runtime provisioning dynamic

## Data Ownership (Rule)
- Setiap modul “memiliki” data/invariants-nya.
- Dilarang join/random query lintas modul di layer domain/usecase.
- Interaksi lintas modul melalui `ports`.

## Hosting v1: Static (Launch)
Alur target (ringkas):
1) User create project/site → API (`hosting` usecase)
2) Worker publish artifact ke objectstore (S3/R2) via `platform/objectstore`
3) Worker configure edge (Cloudflare/CloudFront) via `platform/edge`
4) Domain attach/verify → `domains` usecase + worker sync

## Hosting v2: Dynamic (Future)
Tambahan konsep (tanpa implementasi dulu):
- Port `Provisioner` / `RuntimeProvisioner` untuk create/update/delete runtime
- Worker menjadi eksekutor provisioning
- Hosting module tetap abstrak, tidak mengunci ke AWS detail

## Quality Gates (Target 9/10)
- Dependency boundary audit wajib lulus (`make audit`)
- Unit + integration tests minimal untuk auth + postgres (`make check`)
- Docs wajib punya:
  - README index
  - Architecture peta sistem
  - Threat model dasar
  - ADR untuk setiap keputusan besar
