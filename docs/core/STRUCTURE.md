# Structure & Contracts

Dokumen ini adalah pegangan struktur repo dan kontrak antar layer.
Tujuan: gampang dibaca, gampang diaudit, minim efek berantai saat perubahan.

> Hard rules lebih detail ada di `docs/ai-rules/AI_RULES.md`.
> Dokumen ini sengaja lebih ringkas dan fokus pada “peta + kontrak”.

---

## Scope Produk
- Sistem inti: **API + Worker**.
- Web/dashboard berada di repo lain (client) dan hanya consume API.
- Fokus launch awal: **hosting statis**. Dynamic hosting adalah fase berikutnya.

---

## Layout Repo (High-level)

### `cmd/`
Entrypoint binary.
- `cmd/api`: HTTP API (control-plane)
- `cmd/worker`: job async (provisioning/sync/reconcile)
- `cmd/migrate`: migrasi schema

### `internal/modules/<module>/`
Modul bisnis per bounded-context.
Struktur standar:
- `domain/` : entity + invariants (aturan saat ini: **stdlib only**, sesuai audit script)
- `ports/`  : interface untuk dependency (repo, provider, issuer, queue, dll)
- `usecase/`: orchestration business flow (depend ke domain + ports)
- `transport/http/`: HTTP adapter khusus modul (handler + request/response types)
- `store/...` (opsional): adapter storage spesifik modul (contoh: postgres)

### `internal/transport/http/`
HTTP “core” (lintas modul):
- `router/`: register route global + v1/v2 subrouter
- `middleware/`: request_id, access_log, auth, trust, dll
- `presenter/`: format response envelope + sanitasi error

### `internal/platform/`
Adapter vendor/IO (AWS/Cloudflare/DB drivers/queue/objectstore/token, dll).
Rule: platform tidak boleh bergantung pada HTTP layer.

### Lainnya
- `internal/config/`: load & validate konfigurasi
- `internal/app/bootstrap/`: wiring/DI (wire)
- `migrations/`: SQL schema changes
- `deploy/`: docker/systemd/infra
- `scripts/`: tooling (boundary audit, dsb)

---

## Contracts Antar Layer (Wajib)

### Domain (`internal/modules/*/domain`)
- Fokus: aturan bisnis paling inti (invariants), value/entity.
- Aturan saat ini (mengikuti audit script): **hanya standard library**.
- Dilarang: import `internal/...` dan third-party.
- Catatan: kalau nanti domain butuh shared util, ubah aturan via ADR + update audit script.

### Ports (`internal/modules/*/ports`)
- Berisi interface dependency.
- Boleh import: stdlib + domain (modul sendiri) + tipe kecil yang netral bila diperlukan.

### Usecase (`internal/modules/*/usecase`)
- Orchestrator flow: validate ringan, call ports, bentuk output, return error terstandar.
- Boleh import: domain + ports + shared util netral.
- Dilarang: import `internal/transport/http/...` dan `internal/platform/...`.

### Transport HTTP Modul (`internal/modules/*/transport/http`)
- Tugas: mapping request/response + validasi ringan + panggil usecase.
- Dilarang: import `internal/platform/...` (akses vendor/IO harus via usecase → ports).

### Core HTTP (`internal/transport/http/...`)
- Router/middleware/presenter bersifat lintas modul.
- Dilarang: import `internal/platform/...`.

### Platform (`internal/platform/...`)
- Implementasi nyata untuk ports (DB, queue, objectstore, edge, token, dll).
- Dilarang: import HTTP core / module transport / app bootstrap.

---

## Cara Menambah Endpoint v1 (Pattern)
1) Tambah route di `internal/transport/http/router/v1/<area>/routes.go`
2) Handler ada di `internal/modules/<module>/transport/http/handler.go`
3) Handler memanggil `usecase` (bukan langsung store/platform)
4) Response selalu lewat `presenter/*` (envelope konsisten)

---

## Testing (Ringkas)
- Definisi & build tags: lihat `docs/TESTING.md`
- Unit: default (`go test ./...`)
- Component: `-tags=component` (HTTP in-memory)
- Integration: `-tags=integration` (dependency real)

---

## Docs Rules (Supaya Docs Tidak Membusuk)
- Keputusan arsitektur/perilaku besar → buat ADR di `docs/adr/`
- `docs/ARCHITECTURE.md` adalah peta sistem (high-level, link ke ADR/Threat Model)
- `docs/THREAT_MODEL.md` minimal harus update saat:
  - auth model berubah
  - trust boundary berubah
  - dynamic hosting mulai dikerjakan
