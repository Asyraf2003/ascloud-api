# Structure & Contracts

Dokumen ini adalah pegangan struktur repo dan kontrak antar layer.
Targetnya: repo gampang dibaca, gampang diaudit, dan perubahan tidak menjalar ke mana-mana.

**Audience**
- Contributor repo (kamu sekarang, kamu di masa depan, dan reviewer yang sok tegas).
- Bukan untuk end-user produk.

**Referensi aturan detail**
- Hard rules: `docs/internal/ai/AI_RULES.md`
- Peta sistem: `docs/core/ARCHITECTURE.md`
- Threat model: `docs/core/THREAT_MODEL.md`
- Keputusan arsitektur: `docs/adr/*`

---

## Scope Produk
- Sistem inti: **API + Worker**.
- Web/dashboard berada di repo lain dan hanya consume API.
- Launch awal: **hosting statis**. Dynamic hosting adalah fase berikutnya.

---

## Layout Repo (Ringkas)

### `cmd/`
Entrypoint binary.
- `cmd/api`: HTTP API (control-plane)
- `cmd/worker`: job async (provisioning/sync/reconcile)
- `cmd/migrate`: migrasi schema

### `internal/modules/<module>/`
Modul bisnis per bounded-context.

Struktur standar:
- `domain/` : entity + invariants + semantic errors
- `ports/`  : interface dependency (repo/provider/issuer/queue/etc)
- `usecase/`: orchestration flow (depend ke domain + ports)
- `transport/http/`: HTTP adapter khusus modul (handler + request/response types)
- `store/...` (opsional): adapter storage spesifik modul (misal: postgres)

### `internal/transport/http/`
HTTP “core” (lintas modul):
- `router/`: register route global + v1/v2 subrouter
- `middleware/`: request_id, access_log, auth, trust, dll
- `presenter/`: response envelope + sanitasi error + redaction log

### `internal/platform/`
Adapter vendor/IO (DB driver, queue, objectstore, edge, token, dll).
Rule: platform tidak boleh bergantung pada HTTP layer dan tidak boleh tahu usecase.

### Lainnya
- `internal/config/`: load & validate konfigurasi
- `internal/app/bootstrap/`: wiring/DI (wire)
- `migrations/`: SQL schema changes
- `deploy/`: docker/systemd/infra
- `scripts/`: tooling (boundary audit, dsb)

---

## Kontrak Antar Layer (Wajib)

### Domain (`internal/modules/*/domain`)
Fokus: aturan bisnis inti (invariants), value/entity, semantic errors.

Boleh import:
- standard library
- `internal/shared/*` yang **pure** (lihat ADR 0003)

Dilarang import:
- `internal/platform/*`
- `internal/transport/http/*`
- module lain (`internal/modules/*`)
- third-party packages

Enforcement:
- `scripts/audit_boundaries.sh`

### Ports (`internal/modules/*/ports`)
Berisi interface dependency (repo/provider/issuer/etc).

Boleh import:
- stdlib
- domain (modul sendiri)
- `internal/shared/*` (pure)

Catatan:
- Third-party tipe kecil hanya kalau benar-benar perlu dan disepakati policy.
- Default: minimalkan.

### Usecase (`internal/modules/*/usecase`)
Orchestrator flow: validasi ringan, call ports, bentuk output, return error terstandar.

Boleh import:
- domain + ports
- `internal/shared/*` (pure)

Dilarang:
- `internal/transport/http/*`
- `internal/platform/*`

### Transport HTTP Modul (`internal/modules/*/transport/http`)
Tugas: mapping request/response + validasi ringan + panggil usecase.

Dilarang:
- import `internal/platform/*` (akses IO wajib via usecase → ports)

Rules:
- Response harus lewat presenter/envelope (jangan bikin format sendiri).
- Jangan taruh logika bisnis di handler.

### Core HTTP (`internal/transport/http/...`)
Router/middleware/presenter bersifat lintas modul.

Dilarang:
- import `internal/platform/*`

### Platform (`internal/platform/...`)
Implementasi nyata untuk ports (DB, queue, objectstore, edge, token, dll).

Dilarang import:
- `internal/transport/http/*`
- module transport (`internal/modules/*/transport/http/*`)
- `internal/app/*` (bootstrap/wiring)
- `internal/modules/*/usecase/*`

Prefer:
- platform hanya expose constructor + implement interface ports.

---

## Interaksi Lintas Modul (Anti spaghetti)
- Modul tidak boleh “manggil modul lain” lewat import package.
- Kalau butuh lintas modul: definisikan interface di `ports/` (misal `AccountService`, `AuditSink`) dan inject implementasinya di bootstrap.

---

## Data Ownership (Rule)
- Setiap modul “memiliki” data/invariants-nya.
- Dilarang join/random query lintas modul di layer domain/usecase.
- Kalau ada kebutuhan baca data modul lain: buat port explicit (read model/service), bukan query diam-diam.

---

## Cara Menambah Endpoint v1 (Pattern)
1) Tambah route di `internal/transport/http/router/v1/<area>/routes.go`
2) Handler ada di `internal/modules/<module>/transport/http/handler.go`
3) Handler memanggil usecase (bukan langsung store/platform)
4) Response selalu lewat `internal/transport/http/presenter/*`

---

## Testing (Ringkas)
Detail: `docs/core/TESTING.md`
- Unit: default (`go test ./...`)
- Component: `-tags=component` (HTTP in-memory)
- Integration: `-tags=integration` (dependency real)

---

## Quality Gates (Repo Health)
- `make audit` wajib lulus sebelum PR dianggap “waras”:
  - unit + component tests
  - boundary audit
  - lint/vet
  - file size + package consistency checks
