# Structure & Contracts

Dokumen ini adalah pegangan struktur repo dan kontrak antar layer.
Tujuan: gampang dibaca, gampang diaudit, minim efek berantai saat perubahan.

> Hard rules paling detail ada di `docs/internal/ai/AI_RULES.md`.
> Dokumen ini ringkas dan fokus pada “peta + kontrak”.

---

## Scope Produk
- Sistem inti: **API + Worker**.
- Web/dashboard berada di repo lain (client) dan hanya consume API.
- Fokus launch awal: **hosting statis**. Dynamic hosting adalah fase berikutnya.

---

## Layout Repo

### `cmd/`
Entrypoint binary.
- `cmd/api`: HTTP API (control-plane)
- `cmd/worker`: job async (provisioning/sync/reconcile)
- `cmd/migrate`: migrasi schema

### `internal/modules/<module>/`
Modul bisnis per bounded-context.

Struktur standar:
- `domain/` : entity + invariants (aturan bisnis inti)
- `ports/`  : interface dependency (repo, provider, issuer, queue, dll)
- `usecase/`: orchestration flow (depend ke domain + ports)
- `transport/http/`: HTTP adapter khusus modul (handler + request/response types)
- `store/...` (opsional): adapter storage spesifik modul (contoh: postgres)

### `internal/transport/http/`
HTTP “core” (lintas modul):
- `router/`: register route global + v1/v2 subrouter
- `middleware/`: request_id, access_log, auth, trust, dll
- `presenter/`: response envelope + sanitasi error + redaction log

### `internal/platform/`
Adapter vendor/IO (DB driver, queue, objectstore, edge, token, dll).
Rule: platform tidak boleh bergantung pada HTTP layer.

### Lainnya
- `internal/config/`: load & validate konfigurasi
- `internal/app/bootstrap/`: wiring/DI (wire)
- `migrations/`: SQL schema changes
- `deploy/`: docker/systemd/infra
- `scripts/`: tooling (boundary audit, dsb)

---

## Contracts Antar Layer

### Domain (`internal/modules/*/domain`)
- Fokus: aturan bisnis inti (invariants), value/entity, semantic errors.
- Boleh import:
  - standard library
  - `internal/shared/*` yang **pure** (lihat ADR 0003)
- Dilarang import:
  - `internal/platform/*`
  - `internal/transport/http/*`
  - module lain (`internal/modules/*`)
  - third-party packages
- Enforcement: `scripts/audit_boundaries.sh`

### Ports (`internal/modules/*/ports`)
- Berisi interface dependency.
- Boleh import:
  - stdlib
  - domain (modul sendiri)
  - `internal/shared/*` (pure)
  - Third-party tipe kecil bila perlu dan disetujui policy (contoh: `github.com/google/uuid`), tapi default: minimalkan.

### Usecase (`internal/modules/*/usecase`)
- Orchestrator flow: validate ringan, call ports, bentuk output, return error terstandar.
- Boleh import: domain + ports + `internal/shared/*` (pure).
- Dilarang: import `internal/transport/http/*` dan `internal/platform/*`.

### Transport HTTP Modul (`internal/modules/*/transport/http`)
- Tugas: mapping request/response + validasi ringan + panggil usecase.
- Dilarang: import `internal/platform/*` (akses vendor/IO harus via usecase → ports).
- Response harus lewat presenter/envelope (jangan bikin format sendiri).

### Core HTTP (`internal/transport/http/...`)
- Router/middleware/presenter bersifat lintas modul.
- Dilarang: import `internal/platform/*`.

### Platform (`internal/platform/...`)
- Implementasi nyata untuk ports (DB, queue, objectstore, edge, token, dll).
- Dilarang import:
  - `internal/transport/http/*`
  - module transport (`internal/modules/*/transport/http/*`)
  - `internal/app/*` (bootstrap/wiring)
- Prefer: platform hanya expose constructor + implement interface ports.

---

## Cara Menambah Endpoint v1 (Pattern)
1) Tambah route di `internal/transport/http/router/v1/<area>/routes.go`
2) Handler ada di `internal/modules/<module>/transport/http/handler.go`
3) Handler memanggil usecase (bukan langsung store/platform)
4) Response selalu lewat `internal/transport/http/presenter/*`

---

## Repo Audits (Quality Gates)

Untuk menjaga kesehatan kode (*hygiene*) dan kepatuhan kontrak, repo ini menggunakan gate otomatis yang dijalankan secara berkala atau via CI:

1.  **Boundary Audit (`scripts/audit_boundaries.sh`)**
    * **Tujuan:** Memastikan aturan import antar layer (Domain, Ports, Usecase, Platform) tidak dilanggar.
    * **FAIL jika:** Terjadi *circular dependency* atau layer bawah mengimpor layer atas (misal: Domain impor Platform).

2.  **Content Audit (`scripts/content_audit.sh`)**
    * **Tujuan:** Deteksi file "kosong secara makna" (*meaningless*), placeholder yang tertinggal, atau struktur minimalis tanpa isi.
    * **FAIL jika:** * File `.go` hanya berisi deklarasi `package <name>` tanpa kode/logika setelah komentar dibersihkan.
        * File `.md` hanya berisi heading tanpa deskripsi substansial atau mengandung kata kunci placeholder (TBD, TODO, dll).
    * **Pengecualian:** Folder kosong wajib menggunakan `.gitkeep` atau `README.md` yang deskriptif, bukan file Go kosong.

3.  **Standard Go Gates**
    * `make audit`: Menjalankan gabungan `go vet`, `staticcheck`, dan unit testing.

---

## Testing (Ringkas)
Detail: lihat `docs/core/TESTING.md`
- Unit: default (`go test ./...`)
- Component: `-tags=component` (HTTP in-memory)
- Integration: `-tags=integration` (dependency real)

---

## Docs Rules (Supaya Tidak Membusuk)
- Keputusan besar arsitektur/perilaku → buat ADR di `docs/adr/`
- `docs/core/ARCHITECTURE.md` adalah peta sistem (high-level, link ke ADR/Threat Model)
- `docs/core/THREAT_MODEL.md` minimal update saat:
  - auth model berubah
  - trust boundary berubah
  - dynamic hosting mulai dikerjakan

---

## Repo Hygiene & Audits (Wajib Hijau)

Repo ini punya audit untuk mencegah “struktur palsu” dan efek berantai.

### Audit scripts
- `make audit` (gabungan)
- `bash scripts/docs_audit.sh` (docs wajib lengkap + link valid + ADR format)
- `bash scripts/testtags_audit.sh` (tag test vs nama file konsisten)
- `bash scripts/content_audit.sh` (anti file kosong/meaningless)

### Rules
- Dilarang commit file Go yang hanya berisi `package xxx`.
  - Jika folder belum dipakai: gunakan `.gitkeep`.
  - Jika package sudah sah: gunakan `doc.go` (package comment) dan/atau kontrak minimal.
